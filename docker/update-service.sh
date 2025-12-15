#!/usr/bin/bash

PROJECT_DIR="$HOME/unipilot"
DOCKER_DIR="$PROJECT_DIR/docker"
SERVICE=$1
IMAGE="unipilot-$SERVICE:latest"
TAG_IMAGE="192.168.86.22:5000/unipilot/$SERVICE:latest "
BUILD="Dockerfile.$SERVICE"
COMPRESS="$IMAGE.tar.gz"
REPLICAS=${2:-1}

if [ -z $SERVICE ]; then
	echo "ERROR: Enter at least one argument"
	exit 1
fi

# build the new image
docker build -f "$DOCKER_DIR/$BUILD" -t $IMAGE . &&

# Save the new IMAGE
docker tag $IMAGE $TAG_IMAGE && 

# Push the new image to repository
docker push $TAG_IMAGE

# Update the service
docker service update --image $TAG_IMAGE --force "unipilot_$SERVICE" &&

# Rescale the service
docker service scale "unipilot_$SERVICE"=$REPLICAS &&

# Review the service
docker service ps "unipilot_$SERVICE" &&

# Clean up
docker rmi $IMAGE
