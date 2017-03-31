job('prometheus-operator-unit-tests') {
    concurrentBuild()

    parameters {
        stringParam('sha1')
    }

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
            extensions {
                commitStatus {
                    context('prometheus-operator-unit-tests')
                    triggeredStatus('Tests triggered')
                    startedStatus('Tests started')
                    completedStatus('SUCCESS', 'Success')
                    completedStatus('FAILURE', 'Failure')
                    completedStatus('PENDING', 'Pending')
                    completedStatus('ERROR', 'Error')
                }
            }
        }
    }

    steps {
        shell('docker run --rm -v $PWD:/go/src/github.com/coreos/prometheus-operator -w /go/src/github.com/coreos/prometheus-operator golang make test')
    }
}
job('prometheus-operator-generate-content') {
    concurrentBuild()

    parameters {
        stringParam('sha1')
    }


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
            extensions {
                commitStatus {
                    context('prometheus-operator-docs')
                    triggeredStatus('Tests triggered')
                    startedStatus('Tests started')
                    completedStatus('SUCCESS', 'Success')
                    completedStatus('FAILURE', 'Failure')
                    completedStatus('PENDING', 'Pending')
                    completedStatus('ERROR', 'Error')
                }
            }
        }
    }

    steps {
        shell('docker run -v $PWD:/go/src/github.com/coreos/prometheus-operator -w /go/src/github.com/coreos/prometheus-operator/ golang make generate')
        shell('git diff --exit-code')
    }
}
job('prometheus-operator-e2e-tests') {
    concurrentBuild()

    parameters {
        stringParam('sha1')
    }

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
            extensions {
                commitStatus {
                    context('prometheus-operator-e2e-tests')
                    triggeredStatus('Tests triggered')
                    startedStatus('Tests started')
                    completedStatus('SUCCESS', 'Success')
                    completedStatus('FAILURE', 'Failure')
                    completedStatus('PENDING', 'Pending')
                    completedStatus('ERROR', 'Error')
                }
            }
        }
    }

    steps {
        shell('docker build -t cluster-setup-env scripts/jenkins/.')
    }

    steps {
        shell('docker run --rm -v $PWD:$PWD -v /var/run/docker.sock:/var/run/docker.sock cluster-setup-env /bin/bash -c "cd $PWD && make crossbuild"')
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
