job('po-tests-pr') {
    concurrentBuild()

    // logRotator(daysToKeep, numberToKeep)
    logRotator(10, 10)

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
                    context('prometheus-operator-tests')
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
        shell('./scripts/jenkins/check-make-generate.sh')
    }

    steps {
        shell('./scripts/jenkins/make-test.sh')
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
        postBuildScripts {
            archiveArtifacts('build/**/*')
            onlyIfBuildSucceeds(false)
            onlyIfBuildFails(false)
        }
        wsCleanup()
    }
}

job('po-tests-master') {
    concurrentBuild()

    // logRotator(daysToKeep, numberToKeep)
    logRotator(30, 30)

    scm {
        git {
            remote {
                github('coreos/prometheus-operator')
            }
            branch('master')
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
        githubPush()
        gitHubPushTrigger()
        cron('@daily')
        pollSCM{scmpoll_spec('')}
    }

    steps {
        shell('./scripts/jenkins/check-make-generate.sh')
    }

    steps {
        shell('./scripts/jenkins/make-test.sh')
    }

    steps {
        shell('./scripts/jenkins/run-e2e-tests.sh')
    }

    publishers {
        postBuildScripts {
            steps {
                shell('./scripts/jenkins/push-to-quay.sh')
            }
            onlyIfBuildSucceeds(true)
        }
        postBuildScripts {
            steps {
                shell('./scripts/jenkins/post-e2e-tests.sh')
            }
            onlyIfBuildSucceeds(false)
            onlyIfBuildFails(false)
        }
        postBuildScripts {
            archiveArtifacts('build/**/*')
            onlyIfBuildSucceeds(false)
            onlyIfBuildFails(false)
        }
        slackNotifier {
            room('#team-monitoring')
            teamDomain('coreos')
            authTokenCredentialId('team-monitoring-slack-jenkins')
            notifyFailure(true)
            notifyRegression(true)
            notifyRepeatedFailure(true)
        }
        wsCleanup()
    }
}
