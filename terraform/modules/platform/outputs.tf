output "monitoring_namespace" {
  description = "Namespace of Prometheus and Grafana."
  value       = kubernetes_namespace_v1.monitoring.metadata[0].name
}

output "ingress_namespace" {
  description = "Namespace of the Ingress controller."
  value       = kubernetes_namespace_v1.ingress.metadata[0].name
}
