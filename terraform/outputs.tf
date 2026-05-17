output "instance_public_ip" {
  description = "Public IP address of the deployed EC2 instance"
  value       = aws_instance.app_server.public_ip
}

output "instance_public_dns" {
  description = "Public DNS name of the deployed EC2 instance"
  value       = aws_instance.app_server.public_dns
}

output "app_url" {
  description = "URL to access the application gateway"
  value       = "http://${aws_instance.app_server.public_ip}:8080"
}

output "grafana_url" {
  description = "URL to access Grafana dashboards"
  value       = "http://${aws_instance.app_server.public_ip}:3000"
}

output "prometheus_url" {
  description = "URL to access Prometheus"
  value       = "http://${aws_instance.app_server.public_ip}:9090"
}

output "ssh_command" {
  description = "SSH command to connect to the instance"
  value       = "ssh -i ~/.ssh/${var.key_name}.pem ubuntu@${aws_instance.app_server.public_ip}"
}
