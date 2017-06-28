// The e-mail address used to login as the admin user to the Tectonic Console.
// 
// Note: This field MUST be set manually prior to creating the cluster.
tectonic_admin_email = "monitoring@coreos.com"

// The bcrypt hash of admin user password to login to the Tectonic Console.
// Use the bcrypt-hash tool (https://github.com/coreos/bcrypt-tool/releases/tag/v1.0.0) to generate it.
// 
// Note: This field MUST be set manually prior to creating the cluster.
tectonic_admin_password_hash = ""

// (optional) Extra AWS tags to be applied to created autoscaling group resources.
// This is a list of maps having the keys `key`, `value` and `propagate_at_launch`.
// 
// Example: `[ { key = "foo", value = "bar", propagate_at_launch = true } ]`
// tectonic_autoscaling_group_extra_tags = ""

// Number of Availability Zones your EC2 instances will be deployed across.
// This should be less than or equal to the total number available in the region. 
// Be aware that some regions only have 2.
// If set worker and master subnet CIDRs are calculated automatically.
// 
// Note:
// This field MUST be set manually prior to creating the cluster.
// It MUST NOT be set if availability zones CIDRs are configured using `tectonic_aws_master_custom_subnets` and `tectonic_aws_worker_custom_subnets`.
tectonic_aws_az_count = "2"

// Instance size for the etcd node(s). Example: `t2.medium`.
tectonic_aws_etcd_ec2_type = "t2.medium"

// The amount of provisioned IOPS for the root block device of etcd nodes.
tectonic_aws_etcd_root_volume_iops = "100"

// The size of the volume in gigabytes for the root block device of etcd nodes.
tectonic_aws_etcd_root_volume_size = "30"

// The type of volume for the root block device of etcd nodes.
tectonic_aws_etcd_root_volume_type = "gp2"

// (optional) List of subnet IDs within an existing VPC to deploy master nodes into.
// Required to use an existing VPC and the list must match the AZ count.
// 
// Example: `["subnet-111111", "subnet-222222", "subnet-333333"]`
// tectonic_aws_external_master_subnet_ids = ""

// (optional) ID of an existing VPC to launch nodes into.
// If unset a new VPC is created.
// 
// Example: `vpc-123456`
// tectonic_aws_external_vpc_id = ""

// If set to true, create public facing ingress resources (ELB, A-records).
// If set to false, a "private" cluster will be created with an internal ELB only.
tectonic_aws_external_vpc_public = true

// (optional) List of subnet IDs within an existing VPC to deploy worker nodes into.
// Required to use an existing VPC and the list must match the AZ count.
// 
// Example: `["subnet-111111", "subnet-222222", "subnet-333333"]`
// tectonic_aws_external_worker_subnet_ids = ""

// (optional) Extra AWS tags to be applied to created resources.
// tectonic_aws_extra_tags = ""

// (optional) This configures master availability zones and their corresponding subnet CIDRs directly.
// 
// Example:
// `{ eu-west-1a = "10.0.0.0/20", eu-west-1b = "10.0.16.0/20" }`
// 
// Note that `tectonic_aws_az_count` must be unset if this is specified.
// tectonic_aws_master_custom_subnets = ""

// Instance size for the master node(s). Example: `t2.medium`.
tectonic_aws_master_ec2_type = "t2.medium"

// The amount of provisioned IOPS for the root block device of master nodes.
tectonic_aws_master_root_volume_iops = "100"

// The size of the volume in gigabytes for the root block device of master nodes.
tectonic_aws_master_root_volume_size = "30"

// The type of volume for the root block device of master nodes.
tectonic_aws_master_root_volume_type = "gp2"

// The target AWS region for the cluster.
tectonic_aws_region = "eu-west-2"

// Name of an SSH key located within the AWS region. Example: coreos-user.
tectonic_aws_ssh_key = "jenkins-tpo-ssh-key"

// Block of IP addresses used by the VPC.
// This should not overlap with any other networks, such as a private datacenter connected via Direct Connect.
tectonic_aws_vpc_cidr_block = "10.0.0.0/16"

// (optional) This configures worker availability zones and their corresponding subnet CIDRs directly.
// 
// Example: `{ eu-west-1a = "10.0.64.0/20", eu-west-1b = "10.0.80.0/20" }`
// 
// Note that `tectonic_aws_az_count` must be unset if this is specified.
// tectonic_aws_worker_custom_subnets = ""

// Instance size for the worker node(s). Example: `t2.medium`.
tectonic_aws_worker_ec2_type = "t2.medium"

// The amount of provisioned IOPS for the root block device of worker nodes.
tectonic_aws_worker_root_volume_iops = "100"

// The size of the volume in gigabytes for the root block device of worker nodes.
tectonic_aws_worker_root_volume_size = "30"

// The type of volume for the root block device of worker nodes.
tectonic_aws_worker_root_volume_type = "gp2"

// The base DNS domain of the cluster.
// 
// Example: `openstack.dev.coreos.systems`.
// 
// Note: This field MUST be set manually prior to creating the cluster.
// This applies only to cloud platforms.
tectonic_base_domain = "dev.coreos.systems"

// (optional) The content of the PEM-encoded CA certificate, used to generate Tectonic Console's server certificate.
// If left blank, a CA certificate will be automatically generated.
// tectonic_ca_cert = ""

// (optional) The content of the PEM-encoded CA key, used to generate Tectonic Console's server certificate.
// This field is mandatory if `tectonic_ca_cert` is set.
// tectonic_ca_key = ""

// (optional) The algorithm used to generate tectonic_ca_key.
// The default value is currently recommend.
// This field is mandatory if `tectonic_ca_cert` is set.
// tectonic_ca_key_alg = "RSA"

// The Container Linux update channel.
// 
// Examples: `stable`, `beta`, `alpha`
tectonic_cl_channel = "stable"

// This declares the IP range to assign Kubernetes pod IPs in CIDR notation.
tectonic_cluster_cidr = "10.2.0.0/16"

// The name of the cluster.
// If used in a cloud-environment, this will be prepended to `tectonic_base_domain` resulting in the URL to the Tectonic console.
// 
// Note: This field MUST be set manually prior to creating the cluster.
// Set via env variable
//tectonic_cluster_name = ""

// (optional) DNS prefix used to construct the console and API server endpoints.
// tectonic_dns_name = ""

// (optional) The path of the file containing the CA certificate for TLS communication with etcd.
// 
// Note: This works only when used in conjunction with an external etcd cluster.
// If set, the variables `tectonic_etcd_servers`, `tectonic_etcd_client_cert_path`, and `tectonic_etcd_client_key_path` must also be set.
// tectonic_etcd_ca_cert_path = ""

// (optional) The path of the file containing the client certificate for TLS communication with etcd.
// 
// Note: This works only when used in conjunction with an external etcd cluster.
// If set, the variables `tectonic_etcd_servers`, `tectonic_etcd_ca_cert_path`, and `tectonic_etcd_client_key_path` must also be set.
// tectonic_etcd_client_cert_path = ""

// (optional) The path of the file containing the client key for TLS communication with etcd.
// 
// Note: This works only when used in conjunction with an external etcd cluster.
// If set, the variables `tectonic_etcd_servers`, `tectonic_etcd_ca_cert_path`, and `tectonic_etcd_client_cert_path` must also be set.
// tectonic_etcd_client_key_path = ""

// The number of etcd nodes to be created.
// If set to zero, the count of etcd nodes will be determined automatically.
// 
// Note: This is currently only supported on AWS.
tectonic_etcd_count = "0"

// (optional) List of external etcd v3 servers to connect with (hostnames/IPs only).
// Needs to be set if using an external etcd cluster.
// 
// Example: `["etcd1", "etcd2", "etcd3"]`
// tectonic_etcd_servers = ""

// If set to true, experimental Tectonic assets are being deployed.
tectonic_experimental = false

// The Kubernetes service IP used to reach kube-apiserver inside the cluster
// as returned by `kubectl -n default get service kubernetes`.
tectonic_kube_apiserver_service_ip = "10.3.0.1"

// The Kubernetes service IP used to reach kube-dns inside the cluster
// as returned by `kubectl -n kube-system get service kube-dns`.
tectonic_kube_dns_service_ip = "10.3.0.10"

// The Kubernetes service IP used to reach self-hosted etcd inside the cluster
// as returned by `kubectl -n kube-system get service etcd-service`.
tectonic_kube_etcd_service_ip = "10.3.0.15"

// The path to the tectonic licence file.
// 
// Note: This field MUST be set manually prior to creating the cluster.
tectonic_license_path = "/go/src/github.com/coreos/tectonic-installer/license"

// The number of master nodes to be created.
// This applies only to cloud platforms.
tectonic_master_count = "1"

// The path the pull secret file in JSON format.
// 
// Note: This field MUST be set manually prior to creating the cluster.
tectonic_pull_secret_path = "/go/src/github.com/coreos/tectonic-installer/secret"

// This declares the IP range to assign Kubernetes service cluster IPs in CIDR notation.
tectonic_service_cidr = "10.3.0.0/16"

// If set to true, a vanilla Kubernetes cluster will be deployed, omitting any Tectonic assets.
tectonic_vanilla_k8s = true

// The number of worker nodes to be created.
// This applies only to cloud platforms.
tectonic_worker_count = "3"

tectonic_autoscaling_group_extra_tags = [
  { key = "createdBy", value = "team-monitoring@coreos.com", propagate_at_launch = true },
  { key = "expirationDate", value = "2017-01-01", propagate_at_launch = true }
]

tectonic_aws_extra_tags = {
  "createdBy"="team-monitoring@coreos.com",
  "expirationDate"="2017-01-01"
}
