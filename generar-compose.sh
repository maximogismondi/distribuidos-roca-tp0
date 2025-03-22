#!/bin/bash
echo "This script will generate a docker-compose file for you"
echo "The output file name will be $1"
echo "The number of clients will be $2"

python3 docker_compose_generator.py $1 $2