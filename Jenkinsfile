
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

        stage('Validate Go Formatting') {
            steps {
                bat '''
                    powershell -NoProfile -Command "$files = gofmt -l .; if ($files) { $files; exit 1 } else { Write-Host 'Go formatting is valid' }"
                '''
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
    }

    post {
        success {
            echo 'CI validation completed successfully.'
        }
        failure {
            echo 'CI failed. Check the stage logs for details.'
        }
    }
}
