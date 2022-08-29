pipeline {
    agent any
    tools {
        go 'go1.19'
    }
    environment {
        GO114MODULE = 'on'
        CGO_ENABLED = 0 
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
    }
    stages {
        stage("BUILD") {
            steps {
                echo 'BUILD EXECUTION STARTED'
                sh 'go version'
                sh 'go get ./...'
                sh 'docker build . -t pranav18vk/go-movies-crud'
            }
        }
        stage("HEROKU DEPLOYMENT"){
            steps {
                echo 'DEPLOYING SERVER'
                sh 'git config --global user.email "pranavmasekar4@gmail.com"'
                sh 'git config --global user.name "pranav"'
                sh 'git push heroku master'
            }
        }
    }
}