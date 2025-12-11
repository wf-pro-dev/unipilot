#!/usr/bin/bash

PROJECT_DIR="$HOME/unipilot"
DOCKER_DIR="$PROJECT_DIR/docker"
SERVICE=$1
IMAGE="unipilot-$SERVICE:latest"
BUILD="Dockerfile.$SERVICE"
COMPRESS="$IMAGE.tar.gz"

if [ -z $SERVICE ]; then
	echo "ERROR: Enter at least one argument"
	exit 1
fi

# build the new image
docker build -f "$DOCKER_DIR/$BUILD" -t $IMAGE . &&

# Save the new IMAGE
docker save $IMAGE | gzip > "/tmp/$COMPRESS" &&

# Send to other node
scp "/tmp/$COMPRESS" will@unipilot-2:/tmp/ &&

# Load the new image
ssh will@unipilot-2 "docker load < /tmp/$COMPRESS && rm /tmp/$COMPRESS" && 

# Updtae the docker service
docker service update --force "unipilot_$SERVICE" &&

# Clean UP
rm "/tmp/$COMPRESS"


