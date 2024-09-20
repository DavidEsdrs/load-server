#!/bin/bash

# Variáveis
X=$1 # Número de requisições concorrentes
Y=$2 # Número de vezes para repetir as requisições
URL="http://localhost:3000" # URL do servidor

# Função para fazer a requisição
make_request() {
  curl -s "$URL"
  echo -e "\n"
}

# Loop para repetir Y vezes
for ((i=1; i<=Y; i++)); do
  echo "Rodada $i de $Y"
  
  # Executa X requisições em paralelo
  for ((j=1; j<=X; j++)); do
    make_request &
  done

  # Aguarda todas as requisições terminarem
  wait
done

echo "Todas as requisições foram concluídas."
