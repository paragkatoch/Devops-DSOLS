
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

echo "Restarting all deployments to ensure latest configs/images are applied..."
kubectl rollout restart deployment -n devops-dsols

echo "Waiting for 10 seconds before restarting deployments..."
sleep 10

echo "Waiting for RabbitMQ rollout to complete (only 1 pod running)..."
while true; do
  RABBITMQ_PODS=$(kubectl get pods -n devops-dsols -l app=rabbitmq --no-headers 2>/dev/null | wc -l | tr -d ' ')
  if [ "$RABBITMQ_PODS" -eq 1 ]; then
    echo "RabbitMQ is stable (1 pod)."
    break
  fi
  echo "Currently $RABBITMQ_PODS RabbitMQ pods. Waiting 5s..."
  sleep 5
done

echo "Performing zero-downtime rollout of applications to ensure they connect to the stable RabbitMQ..."
kubectl rollout restart deployment user-controller order-controller product-controller user-service order-service product-service -n devops-dsols

echo "Deployment successful!"
kubectl get pods -n devops-dsols
