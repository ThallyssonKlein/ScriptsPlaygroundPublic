package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

const githubAPI = "https://api.github.com/users/%s/starred"

type Repository struct {
    Name        string `json:"name"`
    FullName    string `json:"full_name"`
    Description string `json:"description"`
    HTMLURL     string `json:"html_url"`
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <github-username>")
        return
    }

    username := os.Args[1]
    url := fmt.Sprintf(githubAPI, username)

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

    for _, repo := range repos {
        fmt.Printf("Name: %s\nFull Name: %s\nDescription: %s\nURL: %s\n\n", repo.Name, repo.FullName, repo.Description, repo.HTMLURL)
    }
}