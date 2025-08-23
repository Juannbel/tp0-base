#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Uso: $0 <nombre_archivo_salida> <cantidad_clientes>"
    exit 1
fi

echo "name: tp0" > $1

echo "
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    volumes:
      - ./server/config.ini:/server/config.ini
    networks:
      - testing_net" >> $1

for i in $(seq 1 $2); do
    echo "
  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - CLI_LOG_LEVEL=DEBUG
    volumes:
      - ./client/config.yaml:/client/config.yaml
    networks:
      - testing_net
    depends_on:
      - server" >> $1
done

echo "
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24" >> $1

echo "Docker compose con $2 clientes generado correctamente en $1"
