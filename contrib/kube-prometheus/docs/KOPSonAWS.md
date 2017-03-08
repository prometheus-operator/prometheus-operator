# Adding kube-prometheus to [KOPS](https://github.com/kubernetes/kops) on AWS 1.5.x


## Prerequisites

A running Kubernetes cluster created with [KOPS](https://github.com/kubernetes/kops).
 
These instructions have currently been tested with  **topology=public** on AWS with KOPS 1.5.1 and Kubernetes 1.5.x

## Open AWS Security Groups:
1. Open port 9100 on the masters security group to the nodes security group
1. Open ports 10250-10252 on the masters security group to the nodes security group.

Example script below requires $AWS\_DEFAULT_PROFILE and [$NAME](https://github.com/kubernetes/kops/blob/master/docs/aws.md#prepare-local-environment)

```bash
MASTER_SG=$(aws --profile ${AWS_DEFAULT_PROFILE} ec2 describe-security-groups --filters "Name=tag:Name,Values=masters.$NAME" --query "SecurityGroups[*].GroupId[]" --output=text)
NODES_SG=$(aws --profile ${AWS_DEFAULT_PROFILE} ec2 describe-security-groups --filters "Name=tag:Name,Values=nodes.$NAME" --query "SecurityGroups[*].GroupId[]" --output=text)
aws --profile ${AWS_DEFAULT_PROFILE} ec2 authorize-security-group-ingress --group-id $MASTER_SG --protocol tcp --port 9100 --source-group $NODES_SG
aws --profile ${AWS_DEFAULT_PROFILE} ec2 authorize-security-group-ingress --group-id $MASTER_SG --protocol tcp --port 10250-10252 --source-group $NODES_SG
```

## Adding kube-prometheus
Following the instructions in the [README](https://github.com/coreos/prometheus-operator/blob/master/contrib/kube-prometheus/README.md):

Example:

```bash
git clone -b master https://github.com/coreos/prometheus-operator.git prometheus-operator-temp;
cd prometheus-operator-temp/contrib/kube-prometheus
./hack/cluster-monitoring/deploy
kubectl -n kube-system create -f manifests/k8s/self-hosted/
cd -
rm -rf prometheus-operator-temp
```
