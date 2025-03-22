#!/bin/bash

# Generar el archivo docker-compose
DOCKER_COMPOSE_FILE="docker-compose-test.yaml"
python3 docker_compose_server_validator.py $DOCKER_COMPOSE_FILE

# Build the images
docker compose -f $DOCKER_COMPOSE_FILE build

# Levantar solo el server en segundo plano
docker compose -f $DOCKER_COMPOSE_FILE up -d server

# Ejecutar el test desde el cliente y capturar el resultado real
docker compose -f $DOCKER_COMPOSE_FILE run --rm test_client
RESULT=$?

# Verificar el resultado y mostrar el mensaje correspondiente
if [ $RESULT -eq 0 ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

# Bajar los servicios y limpiar
docker compose -f $DOCKER_COMPOSE_FILE down

# Eliminar el archivo docker-compose
rm $DOCKER_COMPOSE_FILE
