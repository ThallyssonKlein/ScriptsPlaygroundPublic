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
)

const githubAPI = "https://api.github.com/users/%s/starred?page=%d&per_page=100"

type Repository struct {
    Name        string `json:"name"`
    FullName    string `json:"full_name"`
    Description string `json:"description"`
    HTMLURL     string `json:"html_url"`
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
						"database_id": "14ebca5d3ca88010a1e1d72121cc76cf",
					},
					"properties": map[string]interface{}{
						"Title": map[string]interface{}{
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
				req, err := http.NewRequest("POST", "http://localhost:8080/v1/pages", bytes.NewBuffer(jsonBody))
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
					fmt.Printf("Failed to create page. Status code: %d\n", resp.StatusCode)
					msg.Nack(false, true)
				}
			
				fmt.Println("Page successfully created!")	
				msg.Ack(false)		
			}
		}()
	}

	log.Printf("Waiting for messages. To exit press CTRL+C")
	wg.Wait()
}

func main() {
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

        var repos []Repository
        if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
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
}