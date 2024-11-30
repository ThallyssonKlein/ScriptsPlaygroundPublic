package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "log"

    "github.com/joho/godotenv"
    "github.com/rabbitmq/amqp091-go"
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