variable "grafana_admin_password" {
  description = "Password for the Grafana admin user."
  type        = string
  sensitive   = true
}

variable "grafana_anonymous_viewer" {
  description = "Lets anyone who reaches Grafana view dashboards without logging in. Grafana is never exposed by an Ingress, so reaching it requires cluster access. Meant for test clusters."
  type        = bool
  default     = false
}

variable "prometheus_retention" {
  description = "How long Prometheus keeps metrics. Shorter retention means less disk and memory."
  type        = string
  default     = "12h"
}

variable "chart_versions" {
  description = "Pinned chart versions, so every install is reproducible."
  type = object({
    kube_prometheus_stack = string
    metrics_server        = string
    traefik               = string
  })
  default = {
    kube_prometheus_stack = "92.17.0"
    metrics_server        = "3.14.0"
    traefik               = "42.2.0"
  }
}
