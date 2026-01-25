#!/usr/bin/bash
set -e


SERVICE=${1:-api}
REPLICAS=${2:-1}
TAG=${3:-latest}

PROJECT_DIR="$HOME/unipilot"
DOCKER_DIR="$PROJECT_DIR/docker"
DEV="dev-$SERVICE"
DEV_CONTAINER="$DEV-1"
DEV_IMAGE="$DEV:latest"
IMAGE="unipilot-$SERVICE:latest"
TAG_IMAGE="wwwill-1.lab:5000/unipilot/$SERVICE:$TAG"
BUILD="Dockerfile.$SERVICE"

if [ -z $SERVICE ]; then
	echo "ERROR: Enter at least one argument"
	exit 1
fi

# build the new image
docker build -f "$DOCKER_DIR/$BUILD" -t $IMAGE .

# Save the new IMAGE
docker tag $IMAGE $TAG_IMAGE

# Push the new image to repository
docker push $TAG_IMAGE

# Update the service
docker service update --image $TAG_IMAGE --force "unipilot_$SERVICE"

# Rescale the service
docker service scale "unipilot_$SERVICE"=$REPLICAS

# Review the service
docker service ps "unipilot_$SERVICE"

# Clean up
docker stop $DEV_CONTAINER && docker rm $DEV_CONTAINER

idocker rmi $DEV_IMAGE

docker rmi $IMAGE
