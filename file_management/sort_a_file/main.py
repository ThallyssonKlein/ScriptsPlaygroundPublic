import os
import random

def escolher_arquivo_aleatorio(pasta):
    arquivos = []
    
    # Percorre todas as subpastas recursivamente
    for root, dirs, files in os.walk(pasta):
        for file in files:
            arquivos.append(os.path.join(root, file))
    
    if not arquivos:
        print("A pasta não contém arquivos.")
        return
    
    arquivo_escolhido = random.choice(arquivos)
    
    print(f"Arquivo escolhido: {arquivo_escolhido}")

pasta = '/home/thallyssonklein/MEGA'

escolher_arquivo_aleatorio(pasta)