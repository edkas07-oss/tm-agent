pipeline {
    agent { label 'builder-01' }

    options {
        timeout(time: 15, unit: 'MINUTES')
        ansiColor('xterm')
        disableConcurrentBuilds()
    }

    environment {
        CGO_ENABLED = '0'
    }

    stages {
        stage('Quality Gate 1: Contract & Static Validation') {
            steps {
                sh './scripts/validate.sh'
            }
        }

        stage('Quality Gate 2: Automated Unit & Lifecycle Tests') {
            steps {
                sh './scripts/test.sh'
            }
        }

        stage('Quality Gate 3: Cross-Platform Static Binary Build') {
            steps {
                sh './scripts/build.sh'
            }
        }

        stage('Quality Gate 4: One-Shot Spool Verification') {
            steps {
                sh '''
                    mkdir -p /tmp/ci-agent-spool
                    ./bin/tm-agent --run-once --spool-dir /tmp/ci-agent-spool --target non-existent-test
                    ls -la /tmp/ci-agent-spool/
                    rm -rf /tmp/ci-agent-spool
                '''
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
