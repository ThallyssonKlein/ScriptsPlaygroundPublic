import os

# Diretório onde os arquivos estão localizados (diretório atual)
directory = "."

# Mapeamento dos arquivos para os novos nomes em ordem cronológica
rename_mapping = {
    "transferencia_4.pdf": "01_transferencia.pdf",
    "transferencia_5.pdf": "02_transferencia.pdf",
    "transferencia_6.pdf": "03_transferencia.pdf",
    "transferencia_7.pdf": "04_transferencia.pdf",
    "transferencia_8.pdf": "05_transferencia.pdf",
    "transferencia_9.pdf": "06_transferencia.pdf",
    "transferencia_10.pdf": "07_transferencia.pdf",
    "transferencia_11.pdf": "08_transferencia.pdf",
    "transferencia_12.pdf": "09_transferencia.pdf",
    "transferencia_13.pdf": "10_transferencia.pdf",
    "transferencia_14.pdf": "11_transferencia.pdf",
    "transferencia_15.pdf": "12_transferencia.pdf",
    "transferencia_16.pdf": "13_transferencia.pdf",
    "transferencia_17.pdf": "14_transferencia.pdf",
    "transferencia_18.pdf": "15_transferencia.pdf",
    "transferencia_19.pdf": "16_transferencia.pdf",
    "transferencia_20.pdf": "17_transferencia.pdf",
    "transferencia_21.pdf": "18_transferencia.pdf",
    "transferencia_22.pdf": "19_transferencia.pdf",
    "transferencia_23.pdf": "20_transferencia.pdf",
    "transferencia_25.pdf": "21_transferencia.pdf",
    "transferencia_26.pdf": "22_transferencia.pdf",
    "transferencia_27.pdf": "23_transferencia.pdf",
    "transferencia_28.pdf": "24_transferencia.pdf",
    "transferencia_29.pdf": "25_transferencia.pdf",
    "transferencia_30.pdf": "26_transferencia.pdf",
    "transferencia_32.pdf": "27_transferencia.pdf",
    "transferencia_31.pdf": "28_transferencia.pdf",
    "transferencia_33.pdf": "29_transferencia.pdf",
    "transferencia_34.pdf": "30_transferencia.pdf",
    "transferencia_35.pdf": "31_transferencia.pdf",
    "transferencia_36.pdf": "32_transferencia.pdf",
    "transferencia_37.pdf": "33_transferencia.pdf",
    "transferencia_38.pdf": "34_transferencia.pdf",
    "transferencia_3.pdf": "35_transferencia.pdf",
}

# Listar todos os arquivos no diretório para diagnóstico
print("Arquivos no diretório:")
for filename in os.listdir(directory):
    if filename.endswith(".pdf"):
        print(filename)

# Renomear os arquivos de acordo com o mapeamento
for old_name, new_name in rename_mapping.items():
    old_path = os.path.join(directory, old_name)
    new_path = os.path.join(directory, new_name)
    
    print(f"Verificando {old_path}")
    
    if os.path.exists(old_path):
        os.rename(old_path, new_path)
        print(f"Renamed {old_name} to {new_name}")
    else:
        print(f"{old_name} not found")
 
