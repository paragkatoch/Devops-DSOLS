pipeline {
    agent any

    environment {
        APP_IMAGE = 'paragkatoch/devops-dsols-app:latest'
        PATH = "/opt/homebrew/bin:/usr/local/bin:${env.PATH}"
    }

    stages {
        stage('Verify Environment') {
            steps {
                echo "Verifying Minikube and tools..."
                sh 'minikube status || minikube start'
                sh 'minikube addons enable ingress'
            }
        }

        stage('Build and Push Docker Image') {
            steps {
                echo "Building application image..."
                sh "docker build -t ${APP_IMAGE} -f infra/Dockerfile ."
                
                echo "Pushing image to Docker Hub..."
                withCredentials([usernamePassword(credentialsId: 'dockerhub-credentials', usernameVariable: 'DOCKER_USER', passwordVariable: 'DOCKER_PASS')]) {
                    sh "echo \$DOCKER_PASS | docker login -u \$DOCKER_USER --password-stdin"
                    sh "docker push ${APP_IMAGE}"
                }
            }
        }

        stage('Load Image to Minikube') {
            steps {
                echo "Pulling image from Docker Hub..."
                sh "docker pull ${APP_IMAGE}"

                echo "Loading image directly into Minikube's Docker daemon..."
                sh "minikube image load ${APP_IMAGE}"
            }
        }

        stage('Deploy Configs & Infrastructure') {
            steps {
                echo "Applying ConfigMaps, Postgres, RabbitMQ, Prometheus, and Grafana..."
                sh 'kubectl apply -f k8s/namespace.yaml'
                sh 'kubectl apply -n devops-dsols -f k8s/configmaps.yaml'
                sh 'kubectl apply -n devops-dsols -f k8s/infrastructure/'
            }
        }

        stage('Deploy Applications') {
            steps {
                echo "Restarting deployments to ensure latest code changes are applied..."
                sh 'kubectl rollout restart deployment -n devops-dsols'

                echo "Waiting for RabbitMQ rollout to complete (only 1 pod running)..."
                sh '''
                while true; do
                  RABBITMQ_PODS=$(kubectl get pods -n devops-dsols -l app=rabbitmq --no-headers 2>/dev/null | wc -l | tr -d ' ')
                  if [ "$RABBITMQ_PODS" -eq 1 ]; then
                    echo "RabbitMQ is stable (1 pod)."
                    break
                  fi
                  echo "Currently $RABBITMQ_PODS RabbitMQ pods. Waiting 5s..."
                  sleep 5
                done
                '''
                
                echo "Deleting app controllers and services so they recreate and get fresh RabbitMQ connections..."
                sh 'kubectl delete deployment user-controller order-controller product-controller user-service order-service product-service -n devops-dsols || true'
                
                echo "Applying apps again..."
                sh 'kubectl apply -n devops-dsols -f k8s/apps/'
            }
        }
    }

    post {
        success {
            echo "Deployment successful!"
            sh 'kubectl get pods -n devops-dsols'
        }
        failure {
            echo "Deployment failed! Please check the logs."
        }
    }
}
