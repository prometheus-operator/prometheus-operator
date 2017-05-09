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
            allowMembersOfWhitelistedOrgsAsAdmin()
            triggerPhrase('test this please|please test this')
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
            allowMembersOfWhitelistedOrgsAsAdmin()
            triggerPhrase('test this please|please test this')
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
            allowMembersOfWhitelistedOrgsAsAdmin()
            triggerPhrase('test this please|please test this')
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
        shell('./scripts/jenkins/run-e2e-tests.sh')
    }

    publishers {
        postBuildScripts {
            steps {
                shell('./scripts/jenkins/post-e2e-tests.sh')
            }
            onlyIfBuildSucceeds(false)
            onlyIfBuildFails(false)
        }
    }
}
