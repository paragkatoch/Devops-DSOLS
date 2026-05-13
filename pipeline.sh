
# APP_IMAGE = 'devops-dsols-app:latest'
# PATH = "/opt/homebrew/bin:/usr/local/bin:${env.PATH}"


echo "Verifying Minikube and tools..."
minikube status || minikube start
minikube addons enable ingress


echo "Building application image..."
docker build -t devops-dsols-app:latest -f infra/Dockerfile .


echo "Loading image directly into Minikube's Docker daemon..."
minikube image load devops-dsols-app:latest


echo "Applying ConfigMaps, Postgres, RabbitMQ, Prometheus, and Grafana..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -n devops-dsols -f k8s/configmaps.yaml
kubectl apply -n devops-dsols -f k8s/infrastructure/


echo "Deploying Go applications and Ingress..."
kubectl apply -n devops-dsols -f k8s/apps/


# echo "Restarting deployments to ensure latest code changes are applied..."
kubectl rollout restart deployment -n devops-dsols


echo "Deployment successful!"
kubectl get pods -n devops-dsols
