package main

import (
	"context"
	"log"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

func main() {
	// Token de autenticação do GitHub
	// https://github.com/settings/tokens
	/*
	Autenticação:
	Substitua seu-token-de-acesso-pessoal pelo token obtido em Developer Settings > Tokens.
	O token precisa de permissões como repo ou public_repo, dependendo se o repositório é privado ou público.

	*/
	token := ""

	// Contexto para as operações da API
	ctx := context.Background()

	// Configurando o cliente com autenticação
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Nome do usuário e repositório que você quer "unstar"
	owner := "henrywhitaker3"
	repo := "Speedtest-Tracker"

	// Executa o "unstar"
	_, err := client.Activity.Unstar(ctx, owner, repo)
	if err != nil {
		log.Fatalf("Erro ao remover estrela do repositório: %v", err)
	}

	log.Printf("Estrela removida com sucesso de %s/%s!", owner, repo)
}
