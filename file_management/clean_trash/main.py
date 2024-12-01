import os
import time
from datetime import datetime, timedelta

# Caminho da pasta que você quer limpar
folder_path = '/home/thallyssonklein/.local/share/Trash/files'

# Calcula a data limite (1 mês atrás)
limit_date = datetime.now() - timedelta(days=30)

# Itera sobre todos os arquivos na pasta
for filename in os.listdir(folder_path):
    file_path = os.path.join(folder_path, filename)
    
    # Verifica se é um arquivo
    if os.path.isfile(file_path):
        # Obtém a data de modificação do arquivo
        file_mod_time = datetime.fromtimestamp(os.path.getmtime(file_path))
        
        # Se o arquivo foi modificado há mais de 1 mês, deleta
        if file_mod_time < limit_date:
            os.remove(file_path)
            print(f'{filename} deletado.')