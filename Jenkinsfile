pipeline {
  agent any

  options {
    disableConcurrentBuilds()
  }

  environment {
    FRONTEND_PORT = '8082'
    API_PORT = '8090'
    POSTGRES_PORT = '5433'
    REDIS_PORT = '6380'
    KAFKA_PORT = '9093'
  }

  stages {
    stage('Lint') {
      steps {
        dir('app/frontend') {
          sh 'docker run --rm -v "$PWD":/src -w /src node:22-alpine sh -c "npm ci && npm run lint"'
        }
      }
    }

    stage('Compose Up') {
      steps {
        dir('app') {
          sh 'docker compose down --remove-orphans || true'
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
        sh 'docker compose down --remove-orphans'
      }
    }
  }
}