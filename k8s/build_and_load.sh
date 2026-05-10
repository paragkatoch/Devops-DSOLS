#!/bin/bash
set -e

echo "Building docker image..."
docker build -t devops-dsols-app:latest -f infra/Dockerfile .

echo "Loading image into Minikube..."
minikube image load devops-dsols-app:latest

echo "Done!"
