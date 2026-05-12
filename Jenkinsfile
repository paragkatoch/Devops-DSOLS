pipeline {
    agent any

    environment {
        APP_IMAGE = 'devops-dsols-app:latest'
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

        stage('Build Docker Image') {
            steps {
                echo "Building application image..."
                sh "docker build -t ${APP_IMAGE} -f infra/Dockerfile ."
            }
        }

        stage('Load Image to Minikube') {
            steps {
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
                echo "Deploying Go applications and Ingress..."
                sh 'kubectl apply -n devops-dsols -f k8s/apps/'
                echo "Restarting deployments to ensure latest code changes are applied..."
                sh 'kubectl rollout restart deployment -n devops-dsols'
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
