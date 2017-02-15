job('e2e-tests') {
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
            admins(['mxinden'])
            orgWhitelist(['coreos-inc'])
        }
    }

    steps {
        shell('docker build -t cluster-setup-env scripts/jenkins/.')
    }

    steps {
        shell('docker run --privileged --rm -v $PWD:/go/src/github.com/coreos/prometheus-operator/ cluster-setup-env /bin/bash -c "cd /go/src/github.com/coreos/prometheus-operator && CGO_ENABLED=0 GOOS=linux go build --ldflags \'-extldflags \\"-static\\"\' github.com/coreos/prometheus-operator/cmd/operator"')
    }

    steps {
        shell('mkdir -p .build/linux-amd64/ && cp operator .build/linux-amd64/.')
        shell('docker build -t quay.io/coreos/prometheus-operator-dev:$BUILD_ID .')
        shell('docker login -u="$QUAY_ROBOT_USERNAME" -p="$QUAY_ROBOT_SECRET" quay.io')
        shell('docker push quay.io/coreos/prometheus-operator-dev:$BUILD_ID')
    }

    steps {
        shell('docker run --privileged --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -v $PWD:/go/src/github.com/coreos/prometheus-operator cluster-setup-env /bin/bash -c "cd /go/src/github.com/coreos/prometheus-operator/scripts/jenkins && REPO=quay.io/coreos/prometheus-operator-dev TAG=$BUILD_ID make"')
    }

    publishers {
        postBuildScripts {
            steps {
                shell('docker run --privileged --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -v $PWD:/go/src/github.com/coreos/prometheus-operator cluster-setup-env /bin/bash -c "cd /go/src/github.com/coreos/prometheus-operator/scripts/jenkins && make clean"')
            }
            onlyIfBuildSucceeds(false)
            onlyIfBuildFails(false)
        }
    }
}
