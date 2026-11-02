output "grafana_admin_password" {
  description = "Read it with: terraform output -raw grafana_admin_password"
  value       = random_password.grafana_admin.result
  sensitive   = true
}

output "monitoring_namespace" {
  value = module.platform.monitoring_namespace
}

output "ingress_namespace" {
  value = module.platform.ingress_namespace
}
