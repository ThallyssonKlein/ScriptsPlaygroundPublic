import os
import random

def escolher_arquivo_aleatorio(pasta):
    arquivos = [f for f in os.listdir(pasta) if os.path.isfile(os.path.join(pasta, f))]
    
    if not arquivos:
        print("A pasta não contém arquivos.")
        return
    
    arquivo_escolhido = random.choice(arquivos)
    
    print(f"Arquivo escolhido: {arquivo_escolhido}")

pasta = '/home/thallyssonklein/MEGA'

escolher_arquivo_aleatorio(pasta)