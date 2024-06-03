import os
import shutil

# Diretório de origem e destino
source_dir = '.'
destination_dir = './2023'

# Criar o diretório de destino se não existir
if not os.path.exists(destination_dir):
    os.makedirs(destination_dir)

# Lista de arquivos com datas de 2023
files_2023 = [
    'IMG_20240602_155833359.jpg',
    'IMG_20240602_155923101.jpg',
    'IMG_20240602_155929380.jpg',
    'IMG_20240602_155936090.jpg',
    'IMG_20240602_155942034.jpg',
    'IMG_20240602_155949392.jpg',
    'IMG_20240602_155956186.jpg',
    'IMG_20240602_160025005.jpg',
    'IMG_20240602_160032365.jpg',
    'IMG_20240602_160040482.jpg',
    'IMG_20240602_160046756.jpg',
    'IMG_20240602_160055109.jpg',
    'IMG_20240602_160102924.jpg',
    'IMG_20240602_160111745.jpg',
    'IMG_20240602_160118829.jpg',
    'IMG_20240602_160125405.jpg',
    'IMG_20240602_160131648.jpg',
    'IMG_20240602_160138959.jpg',
    'IMG_20240602_160147586.jpg',
    'IMG_20240602_160153418.jpg',
    'IMG_20240602_160159073.jpg',
    'IMG_20240602_160205224.jpg',
    'IMG_20240602_160211445.jpg',
    'IMG_20240602_160218202.jpg',
    'IMG_20240602_160225452.jpg',
    'IMG_20240602_160231603.jpg'
]

# Mover os arquivos
for file_name in files_2023:
    source_path = os.path.join(source_dir, file_name)
    destination_path = os.path.join(destination_dir, file_name)
    if os.path.exists(source_path):
        shutil.move(source_path, destination_path)
        print(f'Movido: {file_name}')
    else:
        print(f'Arquivo não encontrado: {file_name}')
 
