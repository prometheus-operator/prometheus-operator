output "kops_state_bucket" {
  value = "s3://${aws_s3_bucket.kops-state.id}"
}

output "kops_master_security_group" {
  value = "${aws_security_group.allow_all.id}"
}

output "kops_main_vpc" {
  value = "${aws_vpc.main.id}"
}
