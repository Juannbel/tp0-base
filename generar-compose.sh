#!/bin/bash

DOCUMENTOS=(19782345 29182342 46782642 42783001 22374590)
NOMBRES=("Juan" "Maria" "Pedro" "Ana" "Luis")
APELLIDOS=("Perez" "Gomez" "Lopez" "Martinez" "Hernandez")
NACIMIENTO=("2000-01-01" "1999-05-12" "1998-07-23" "1997-11-30" "1996-03-15")
NUMEROS=("6789" "9321" "8923" "4987" "7891")

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
    volumes:
      - ./server/config.ini:/config.ini
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
      - NOMBRE=${NOMBRES[$i-1]}
      - APELLIDO=${APELLIDOS[$i-1]}
      - DOCUMENTO=${DOCUMENTOS[$i-1]}
      - NACIMIENTO=${NACIMIENTO[$i-1]}
      - NUMERO=${NUMEROS[$i-1]}
    volumes:
      - ./client/config.yaml:/config.yaml
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
