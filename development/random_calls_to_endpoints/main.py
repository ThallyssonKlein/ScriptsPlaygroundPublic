import requests
from random import randint

# Lista de endpoints
endpoints = [
    "/cedente-pessoa-juridica-sucesso",
    "/cedente-pessoa-juridica-falha",
    "/cedente-pessoa-fisica-sucesso",
    "/cedente-pessoa-fisica-falha",
    "/sacado-sucesso",
    "/sacado-falha"
]

# URL base do seu serviço
base_url = "http://localhost:8080"

# Para cada endpoint na lista
for endpoint in endpoints:
    # Gere um número aleatório de vezes para chamar o endpoint
    num_calls = randint(1, 10)

    # Chame o endpoint o número aleatório de vezes
    for _ in range(num_calls):
        response = requests.get(base_url + endpoint)

        # Imprima a resposta
        print(f"Response from {endpoint}: {response.text}")