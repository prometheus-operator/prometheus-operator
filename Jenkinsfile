job('prometheus-operator-unit-tests') {
    concurrentBuild()

    scm {
        git {
            remote {
                github('coreos/prometheus-operator')
                refspec('+refs/pull/*:refs/remotes/origin/pr/*')
            }
            branch('${sha1}')
        }
    }

    triggers {
        githubPullRequest {
            useGitHubHooks()
            orgWhitelist(['coreos-inc'])
        }
    }

    steps {
        shell('docker run --rm -v $PWD:/go/src/github.com/coreos/prometheus-operator -w /go/src/github.com/coreos/prometheus-operator golang make test')
    }
}
job('prometheus-operator-e2e-tests') {
    concurrentBuild()

    scm {
        git {
            remote {
                github('coreos/prometheus-operator')
                refspec('+refs/pull/*:refs/remotes/origin/pr/*')
            }
            branch('${sha1}')
        }
    }

    wrappers {
        credentialsBinding {
            amazonWebServicesCredentialsBinding{
                accessKeyVariable('AWS_ACCESS_KEY_ID')
                secretKeyVariable('AWS_SECRET_ACCESS_KEY')
                credentialsId('Jenkins-Monitoring-AWS-User')
            }
            usernamePassword('QUAY_ROBOT_USERNAME', 'QUAY_ROBOT_SECRET', 'quay_robot')
        }
    }

    triggers {
        githubPullRequest {
            useGitHubHooks()
            orgWhitelist(['coreos-inc'])
        }
    }

    steps {
        shell('docker build -t cluster-setup-env scripts/jenkins/.')
    }

    steps {
        shell('docker run --rm -v /var/jenkins/workspace/e2e-playground:/var/jenkins/workspace/e2e-playground -v /var/run/docker.sock:/var/run/docker.sock cluster-setup-env /bin/bash -c "cd /var/jenkins/workspace/e2e-playground && make crossbuild"')
    }

    steps {
        shell('docker build -t quay.io/coreos/prometheus-operator-dev:$BUILD_ID .')
        shell('docker login -u="$QUAY_ROBOT_USERNAME" -p="$QUAY_ROBOT_SECRET" quay.io')
        shell('docker push quay.io/coreos/prometheus-operator-dev:$BUILD_ID')
    }

    steps {
        shell('docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -v $PWD:/go/src/github.com/coreos/prometheus-operator cluster-setup-env /bin/bash -c "cd /go/src/github.com/coreos/prometheus-operator/scripts/jenkins && REPO=quay.io/coreos/prometheus-operator-dev TAG=$BUILD_ID make"')
    }

    publishers {
        postBuildScripts {
            steps {
                shell('docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -v $PWD:/go/src/github.com/coreos/prometheus-operator cluster-setup-env /bin/bash -c "cd /go/src/github.com/coreos/prometheus-operator/scripts/jenkins && make clean"')
                shell('docker rmi quay.io/coreos/prometheus-operator-dev:$BUILD_ID')
            }
            onlyIfBuildSucceeds(false)
            onlyIfBuildFails(false)
        }
    }
}
