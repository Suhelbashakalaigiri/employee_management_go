
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
