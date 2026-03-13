#!/bin/bash

echo "Compose file: $1"
echo "Amount of clients: $2"

python3 generador.py $1 $2 > docker-compose.yml
