FROM golang:1.7.5-wheezy

ENV TERRAFORM_VERSION 0.8.4
ENV KOPS_VERSION 1.5.1

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates \
		wget \
    unzip \
    python python-pip \
    && rm -rf /var/lib/apt/lists/*

RUN wget -q -O /terraform.zip "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip /terraform.zip -d /bin

RUN wget -q -O /kops "https://github.com/kubernetes/kops/releases/download/${KOPS_VERSION}/kops-linux-amd64" && \
    chmod +x /kops && \
    mv /kops /bin

RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip" && \
    unzip awscli-bundle.zip && \
    ./awscli-bundle/install -i /usr/local/aws -b /bin/aws

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /bin/kubectl

RUN curl -L https://get.docker.com/builds/Linux/x86_64/docker-1.6.0 > /usr/local/bin/docker && \
    chmod +x /usr/local/bin/docker