pipeline {
  stages {
    stage('Prepare') {
      steps {
        sh 'make install'
      }
    }
    stage('Build') {
      parallel {
        stage('Build') {
          steps {
            sh 'make build'
          }
        }
        stage('Test') {
          steps {
            sh 'make test'
          }
        }
      }
    }
    stage('Image') {
      steps {
        sh 'make docker_build'
      }
    }
    stage('Push') {
      steps {
        echo 'Push it!'
      }
    }
  }
}