pipeline {
  agent any

  environment {
    FRONTEND_PORT = '8082'
  }

  stages {
    stage('Compose Up') {
      steps {
        dir('app') {
          sh 'docker compose up -d --build'
        }
      }
    }

    stage('Wait for API') {
      steps {
        sh """#!/bin/bash
          for i in \$(seq 1 30); do
            if curl -sf http://localhost:${env.FRONTEND_PORT}/healthz >/dev/null 2>&1; then
              echo 'API is up'
              exit 0
            fi
            sleep 2
          done
          echo 'API failed to start' >&2
          exit 1
        """
      }
    }

    stage('Smoke Test via Proxy') {
      steps {
        sh "curl -sf http://localhost:${env.FRONTEND_PORT}/api/v1/presets -o /dev/null"
      }
    }
  }

  post {
    always {
      dir('app') {
        sh 'docker compose down'
      }
    }
  }
}