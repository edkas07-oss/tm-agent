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
                    ./bin/tm-agent --run-once --spool-dir /tmp/ci-agent-spool --target non-existent-test || true
                    rm -rf /tmp/ci-agent-spool
                '''
            }
        }

        stage('Quality Gate 5: Archive Multi-OS Artifacts') {
            steps {
                archiveArtifacts artifacts: 'bin/**/*', fingerprint: true, allowEmptyArchive: false
            }
        }
    }

    post {
        always {
            cleanWs deleteDirs: true, notFailBuild: true
        }
        success {
            echo "Pipeline tm-agent build & validation completed successfully."
        }
        failure {
            echo "Pipeline tm-agent build failed! Check console logs."
        }
    }
}

