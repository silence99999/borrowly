variable "aws_region" {
  description = "AWS region to deploy the infrastructure in"
  type        = string
  default     = "eu-west-1"
}

variable "project_name" {
  description = "Project identifier used for resource naming and tagging"
  type        = string
  default     = "rent-items"
}

variable "environment" {
  description = "Deployment environment (e.g. dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "ami_id" {
  description = "AMI ID for the EC2 instance (Ubuntu 22.04 LTS in eu-west-1)"
  type        = string
  default     = "ami-0d64bb532e0502c46"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.small"
}

variable "key_name" {
  description = "Name of the EC2 key pair for SSH access (must already exist in AWS)"
  type        = string
}

variable "disk_size_gb" {
  description = "Root EBS volume size in gigabytes"
  type        = number
  default     = 20
}

variable "admin_cidr_blocks" {
  description = "CIDR blocks allowed to access SSH, Prometheus, and Grafana (restrict to your IP)"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "github_actions_cidr_blocks" {
  description = "GitHub Actions IP ranges allowed to SSH for deployment"
  type        = list(string)
  # Current GitHub Actions IP ranges - update if deployments start timing out
  # Source: https://api.github.com/meta (actions key)
  default = [
    "4.148.0.0/16",
    "20.1.0.0/16",
    "20.7.0.0/16",
    "20.232.0.0/16",
    "20.248.0.0/16",
    "4.175.0.0/16",
    "20.105.0.0/16",
    "20.200.0.0/16",
    "20.201.0.0/16",
    "20.205.0.0/16",
    "20.207.0.0/16",
    "20.233.0.0/16"
  ]
}
