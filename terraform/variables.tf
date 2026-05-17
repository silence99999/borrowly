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
