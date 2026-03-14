#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <output_filename> <client_amount>"
    exit 1
fi

if ! [[ "$2" =~ ^[1-9]+$ ]]; then
    echo "Client amount must be a positive integer."
    exit 1
fi

echo "Compose file: $1"
echo "Amount of clients: $2"

python3 generador.py $2 > $1
