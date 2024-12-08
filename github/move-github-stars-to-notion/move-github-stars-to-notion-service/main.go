package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "log"
    "time"

    "github.com/joho/godotenv"
    "github.com/rabbitmq/amqp091-go"

	"runtime"
	"sync"
	"bytes"

	"move-github-stars-to-notion/proxy"

	"io/ioutil"
	"io"
	"strings"
)

const githubAPI = "http://localhost:8082/users/%s/starred?page=%d&per_page=100"

// Repository structure to unmarshal GitHub API response
type Repository struct {
    Name        string `json:"name"`
    FullName    string `json:"full_name"`
    Description string `json:"description"`
    HTMLURL     string `json:"html_url"`
}

func unstar(repo string, owner string) {
    token := os.Getenv("GITHUB_TOKEN")
    fmt.Println("Token: ", token)

    url := fmt.Sprintf("https://api.github.com/user/starred/%s/%s", owner, repo)

	fmt.Println(url)

    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        fmt.Printf("Erro ao criar a requisição: %v\n", err)
        return
    }

    req.Header.Set("Authorization", "token "+token)
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Erro ao fazer a requisição: %v\n", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNoContent {
        fmt.Println("Estrela removida com sucesso!")
    } else {
        bodyBytes, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Falha ao ler o corpo da resposta: %v\n", err)
            return
        }
        bodyString := string(bodyBytes)
        fmt.Printf("Falha ao remover estrela: %s\n", bodyString)
    }
}

func sendToRabbitMQ(queueName, message string) error {
    rabbitMQURL := os.Getenv("RABBITMQ_URL")
    conn, err := amqp091.Dial(rabbitMQURL)
    if err != nil {
        return fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
    }
    defer conn.Close()

    channel, err := conn.Channel()
    if err != nil {
        return fmt.Errorf("falha ao criar canal: %w", err)
    }
    defer channel.Close()

    _, err = channel.QueueDeclare(
        queueName,
        true,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return fmt.Errorf("falha ao declarar fila: %w", err)
    }

    err = channel.PublishWithContext(
        nil,
        "",
        queueName,
        false,
        false,
        amqp091.Publishing{
            ContentType: "text/plain",
            Body:        []byte(message),
        },
    )
    if err != nil {
        return fmt.Errorf("falha ao publicar mensagem: %w", err)
    }

    return nil
}

func consumeFromRabbitMQ(queueName string) {
    rabbitMQURL := os.Getenv("RABBITMQ_URL")
    conn, err := amqp091.Dial(rabbitMQURL)
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        log.Fatalf("Failed to open a channel: %v", err)
    }
    defer ch.Close()

    err = ch.Qos(1, 0, false)
    if err != nil {
        log.Fatalf("Failed to set QoS: %v", err)
    }

    msgs, err := ch.Consume(
        queueName,
        "",
        false,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        log.Fatalf("Failed to register a consumer: %v", err)
    }

    numCPUs := runtime.NumCPU()
    var wg sync.WaitGroup
    wg.Add(numCPUs)

    for i := 0; i < numCPUs; i++ {
        go func() {
            defer wg.Done()
            for msg := range msgs {
                log.Printf("Received a message: %s", msg.Body)
                repoURL := string(msg.Body)

                requestBody := map[string]interface{}{
                    "parent": map[string]interface{}{
                        "database_id": "156bca5d3ca88009aa49dea20a474c68",
                    },
                    "properties": map[string]interface{}{
                        "Name": map[string]interface{}{
                            "title": []map[string]interface{}{
                                {
                                    "text": map[string]interface{}{
                                        "content": repoURL,
                                    },
                                },
                            },
                        },
                    },
                }

                jsonBody, err := json.Marshal(requestBody)
                if err != nil {
                    fmt.Println("Error marshalling request body:", err)
                    return
                }

                req, err := http.NewRequest("POST", "http://localhost:8083/v1/pages", bytes.NewBuffer(jsonBody))
                if err != nil {
                    fmt.Println("Error creating HTTP request:", err)
                    return
                }

                token := os.Getenv("NOTION_API_TOKEN")
                req.Header.Set("Authorization", "Bearer "+token)
                req.Header.Set("Content-Type", "application/json")
                req.Header.Set("Notion-Version", "2022-06-28")

                client := &http.Client{}
                resp, err := client.Do(req)
                if err != nil {
                    fmt.Println("Error making request:", err)
                    return
                }
                defer resp.Body.Close()

                if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
                    bodyBytes, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        fmt.Println("Error reading response body:", err)
                    } else {
                        bodyString := string(bodyBytes)
                        fmt.Printf("Failed to create page. Status code: %d. Response: %s\n", resp.StatusCode, bodyString)
                    }
                    msg.Nack(false, true)
                } else {
                    fmt.Println("Page successfully created!")
                    parts := strings.Split(repoURL, "/")
                    fmt.Println(parts)
                    owner := parts[3]
                    repo := parts[4]                
                    unstar(repo, owner)
                    fmt.Println("Unstarred repository!")
                    msg.Ack(false)
                }
            }
        }()
    }

    log.Printf("Waiting for messages. To exit press CTRL+C")
    wg.Wait()
}

func fetchRepositoriesWithRetry(url string, maxRetries int) ([]Repository, error) {
    var repos []Repository
    for attempts := 0; attempts < maxRetries; attempts++ {
        resp, err := http.Get(url)
        if err != nil {
            log.Printf("Error fetching data from GitHub (attempt %d): %v", attempts+1, err)
            time.Sleep(2 * time.Second)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            log.Printf("Error: received status code %d (attempt %d)", resp.StatusCode, attempts+1)
            time.Sleep(2 * time.Second)
            continue
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Error reading response body (attempt %d): %v", attempts+1, err)
            time.Sleep(2 * time.Second)
            continue
        }

        if err := json.Unmarshal(body, &repos); err != nil {
            log.Printf("Error decoding JSON response (attempt %d): %v", attempts+1, err)
            time.Sleep(2 * time.Second)
            continue
        }

        return repos, nil
    }
    return nil, fmt.Errorf("falha ao obter repositórios após %d tentativas", maxRetries)
}

func main() {
    go proxy.StartProxy(":8082", "https://api.github.com")
    go proxy.StartProxy(":8083", "https://api.notion.com")

    err2 := godotenv.Load()
    if err2 != nil {
        log.Fatalf("Erro ao carregar o arquivo .env: %v", err2)
    }

    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <github-username>")
        return
    }

    username := os.Args[1]
    page := 1

    for {
        fmt.Println("-------------------")
        url := fmt.Sprintf(githubAPI, username, page)
        repos, err := fetchRepositoriesWithRetry(url, 3) // 3 tentativas por página
        if err != nil {
            log.Fatalf("Erro ao obter os repositórios: %v", err)
        }

        if len(repos) == 0 {
            break
        }

        for _, repo := range repos {
            fmt.Printf("Name: %s\nFull Name: %s\nDescription: %s\nURL: %s\n\n", repo.Name, repo.FullName, repo.Description, repo.HTMLURL)
            err := sendToRabbitMQ("repo_urls", repo.HTMLURL)
            if err != nil {
                log.Printf("Erro ao enviar mensagem ao RabbitMQ: %v", err)
            }
        }

        page++
        fmt.Println("-------------------")
    }

    consumeFromRabbitMQ("repo_urls")
}
