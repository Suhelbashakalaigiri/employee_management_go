
pipeline {
    agent any

    options {
        timestamps()
        disableConcurrentBuilds()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Go Vet') {
            steps {
                bat 'go vet ./...'
            }
        }

        stage('Go Test') {
            steps {
                bat 'go test ./...'
            }
        }

        stage('Go Build') {
            steps {
                bat 'go build ./...'
            }
        }

        stage('Docker Build') {
            steps {
                bat 'docker build -t suhelbasha7324/employee-management-go:latest .'
            }
        }

        stage('Docker Push') {
            steps {
                withCredentials([
                    usernamePassword(
                        credentialsId: 'dockerhub-creds',
                        usernameVariable: 'DOCKER_USER',
                        passwordVariable: 'DOCKER_TOKEN'
                    )
                ]) {
                    bat '''
                        @echo off
                        powershell -NoProfile -Command "$env:DOCKER_TOKEN | docker login -u $env:DOCKER_USER --password-stdin"
                        if errorlevel 1 exit /b 1

                        docker push suhelbasha7324/employee-management-go:latest
                        if errorlevel 1 exit /b 1

                        docker logout
                    '''
                }
            }
        }
    }

    post {
        success {
            echo 'CI pipeline completed successfully.'
        }

        failure {
            echo 'CI pipeline failed. Check the failed stage in Console Output.'
        }

        always {
            echo 'CI pipeline execution finished.'
        }
    }
}
