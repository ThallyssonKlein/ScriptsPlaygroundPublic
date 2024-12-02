package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "log"

    "github.com/joho/godotenv"
    "github.com/rabbitmq/amqp091-go"

	"runtime"
	"sync"
	"bytes"

	"move-github-stars-to-notion/proxy"

	"io/ioutil"
	"io"
	"path"
    "net/url"
)

const githubAPI = "http://localhost:8082/users/%s/starred?page=%d&per_page=100"

type Repository struct {
    Name        string `json:"name"`
    FullName    string `json:"full_name"`
    Description string `json:"description"`
    HTMLURL     string `json:"html_url"`
}

func unstar(repo string) {
		// Defina os valores apropriados
		token := os.Getenv("GITHUB_TOKEN")
		owner := "thallyssonklein"          // Substitua pelo dono do repositório
	
		// URL da API para remover estrela de um repositório específico
		url := fmt.Sprintf("https://api.github.com/user/starred/%s/%s", owner, repo)
	
		// Cria uma nova requisição DELETE
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf("Erro ao criar a requisição: %v\n", err)
			return
		}
	
		// Define o cabeçalho de autenticação
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
	
		// Cria um cliente HTTP para enviar a requisição
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Erro ao fazer a requisição: %v\n", err)
			return
		}
		defer resp.Body.Close()
	
		// Verifica a resposta
		if resp.StatusCode == http.StatusNoContent {
			fmt.Println("Estrela removida com sucesso!")
		} else {
			fmt.Printf("Falha ao remover estrela: %s\n", resp.Status)
		}	
}

func sendToRabbitMQ(queueName, message string) error {
	// Conectar ao RabbitMQ
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	conn, err := amqp091.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
	}
	defer conn.Close()

	// Criar um canal
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("falha ao criar canal: %w", err)
	}
	defer channel.Close()

	// Declarar uma fila
	_, err = channel.QueueDeclare(
		queueName,
		true,  // Durável
		false, // Auto-delete
		false, // Exclusiva
		false, // Sem espera
		nil,   // Argumentos adicionais
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar fila: %w", err)
	}

	// Publicar mensagem na fila
	err = channel.PublishWithContext(
		nil, // Contexto
		"",  // Exchange
		queueName, // Roteamento
		false,     // Obrigatório
		false,     // Imediato
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

	// Definir prefetch count para garantir que apenas uma mensagem seja entregue por vez
	err = ch.Qos(
		1,     // Prefetch count: 1 mensagem por vez
		0,     // Prefetch size: sem limite específico
		false, // Apply to channel
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Consumir a fila usando ch.Consume()
	msgs, err := ch.Consume(
		queueName, // Nome da fila
		"",        // Consumer tag (gerado automaticamente)
		false,     // Auto Ack (não queremos auto-ack, queremos confirmar manualmente)
		false,     // Exclusivo
		false,     // No local
		false,     // No-wait
		nil,       // Argumentos adicionais
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// Consumir mensagens com múltiplas goroutines para processá-las
	numCPUs := runtime.NumCPU() // Número de núcleos da CPU
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
						"database_id": "150bca5d3ca880ca8bc8f9f0b129acc5",
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
			
			
				// Converte a estrutura da requisição para JSON.
				jsonBody, err := json.Marshal(requestBody)
				if err != nil {
					fmt.Println("Error marshalling request body:", err)
					return
				}
			
				// Cria a requisição HTTP.
				req, err := http.NewRequest("POST", "http://localhost:8083/v1/pages", bytes.NewBuffer(jsonBody))
				if err != nil {
					fmt.Println("Error creating HTTP request:", err)
					return
				}
			
				// Define os headers.
				token := os.Getenv("NOTION_API_TOKEN")
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Notion-Version", "2022-06-28") // Ajuste para a versão da API que você está usando.
			
				// Executa a requisição.
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println("Error making request:", err)
					return
				}
				defer resp.Body.Close()

				// Verifica a resposta.
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
					u, err := url.Parse(repoURL)
					if err != nil {
						fmt.Println("Error parsing URL:", err)
						return
					}				
					repoName := path.Base(u.Path)
					unstar(repoName)
					fmt.Println("Unstarred repository!")
					msg.Ack(false)
				}			
			}
		}()
	}

	log.Printf("Waiting for messages. To exit press CTRL+C")
	wg.Wait()
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
        resp, err := http.Get(url)
        if err != nil {
            fmt.Printf("Error fetching data from GitHub: %v\n", err)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            fmt.Printf("Error: received status code %d\n", resp.StatusCode)
            return
        }

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Error reading response body: %v\n", err)
            return
        }

        var repos []Repository
        if err := json.Unmarshal(body, &repos); err != nil {
            fmt.Printf("Error decoding JSON response: %v\n", err)
            return
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