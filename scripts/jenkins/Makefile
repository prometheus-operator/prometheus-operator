CLUSTER_NAME  ?= prom-test-$(shell whoami)
DOMAIN        ?= dev.coreos.systems

path          ?= clusters/${CLUSTER_NAME}
build_path    := $(path)/.build
aws_region     = eu-west-1

KOPS_CMD        = kops --state $(shell terraform output -state "$(build_path)/terraform.tfstate" kops_state_bucket)
TERRAFORM_FLAGS = -var "dns_domain=$(DOMAIN)" -var "cluster_name=$(CLUSTER_NAME)" -state "$(build_path)/terraform.tfstate" 

all: gen-ssh cluster wait-for-cluster run-e2e-tests

clean: clean-cluster clean-aws-deps

gen-ssh:
	ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa -q

aws-deps:
	AWS_REGION=$(aws_region) terraform apply $(TERRAFORM_FLAGS) ./templates

cluster: aws-deps
	$(KOPS_CMD) get cluster | grep -v $(CLUSTER_NAME).$(DOMAIN) || \
	$(KOPS_CMD) create cluster \
		--name $(CLUSTER_NAME).$(DOMAIN) \
		--cloud aws --zones $(aws_region)a --kubernetes-version 1.5.2 \
		--master-size t2.medium --yes \
		--master-security-groups $(shell terraform output -state "$(build_path)/terraform.tfstate" kops_master_security_group) \
		--node-security-groups $(shell terraform output -state "$(build_path)/terraform.tfstate" kops_master_security_group) \
		--vpc $(shell terraform output -state "$(build_path)/terraform.tfstate" kops_main_vpc)
	EDITOR='./ed.sh manifests/kops/regular-ig.yaml' $(KOPS_CMD) edit ig nodes
	$(KOPS_CMD) update cluster --yes

run-e2e-tests:
	$(MAKE) -C ../../ e2e-test

wait-for-cluster:
	timeout 1800 ./wait-for-cluster.sh

clean-cluster:
	$(KOPS_CMD) delete cluster --name $(CLUSTER_NAME).$(DOMAIN) --yes

clean-aws-deps:
	AWS_REGION=$(aws_region) terraform destroy -force $(TERRAFORM_FLAGS) ./templates
	rm -f $(build_path)/terraform.tfstate*
