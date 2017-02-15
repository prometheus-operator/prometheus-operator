variable "dns_domain" {}

variable "cluster_name" {}

data "aws_route53_zone" "monitoring_zone" {
  name = "${var.dns_domain}"
}

resource "aws_route53_zone" "cluster_zone" {
  name = "${var.cluster_name}.${var.dns_domain}"
}

resource "aws_route53_record" "cluster_zone_record" {
  name    = "${var.cluster_name}.${var.dns_domain}"
  zone_id = "${data.aws_route53_zone.monitoring_zone.zone_id}"
  type    = "NS"
  ttl     = "300"
  records = ["${aws_route53_zone.cluster_zone.name_servers}"]
}

resource "aws_s3_bucket" "kops-state" {
  bucket = "kops-${sha1("${var.cluster_name}-${var.dns_domain}")}"
}

resource "aws_security_group" "allow_all" {
    name = "allow_all"
    description = "Allow all inbound traffic"
    vpc_id = "${aws_vpc.main.id}"

    ingress {
        from_port = 30000
        to_port = 32767
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 80
        to_port = 80
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags {
        Name = "allow_all"
     }
}

resource "aws_vpc" "main" {
    cidr_block = "172.20.0.0/16"
}

resource "aws_internet_gateway" "gw" {
    vpc_id = "${aws_vpc.main.id}"
}


