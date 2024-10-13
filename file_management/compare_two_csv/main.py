import pandas as pd

# Carrega os dois CSVs
df1 = pd.read_csv('./csv_1.csv')

import chardet    

rawdata = open('./csv_1.csv', 'rb').read()
result = chardet.detect(rawdata)
encoding = result['encoding']
df2 = pd.read_csv('./csv_2.csv', encoding=encoding)

# Suponha que as colunas são 'X' em df1 e 'Y' em df2
coluna_df1 = df1['Field']
coluna_df2 = df2['Field']

# Encontra os itens que estão apenas no primeiro CSV
itens_somente_df1 = coluna_df1[~coluna_df1.isin(coluna_df2)].dropna()

# Encontra os itens que estão apenas no segundo CSV
itens_somente_df2 = coluna_df2[~coluna_df2.isin(coluna_df1)].dropna()

print("Itens somente no primeiro CSV:")
print(itens_somente_df1)

print("Itens somente no segundo CSV:")
print(itens_somente_df2)
