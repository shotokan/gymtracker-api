# Cluster add-ons the application relies on: metrics for the HPA, an Ingress
# controller, and the Prometheus and Grafana stack.

resource "kubernetes_namespace_v1" "ingress" {
  metadata {
    name = "ingress"
  }
}

resource "kubernetes_namespace_v1" "monitoring" {
  metadata {
    name = "monitoring"
  }
}

resource "helm_release" "metrics_server" {
  name       = "metrics-server"
  namespace  = "kube-system"
  repository = "https://kubernetes-sigs.github.io/metrics-server"
  chart      = "metrics-server"
  version    = var.chart_versions.metrics_server
  wait       = true
  timeout    = 300

  values = [yamlencode({
    args = [
      # minikube and kind serve kubelet metrics with self-signed certificates.
      "--kubelet-insecure-tls",
      "--metric-resolution=15s",
    ]
  })]
}

resource "helm_release" "traefik" {
  name       = "traefik"
  namespace  = kubernetes_namespace_v1.ingress.metadata[0].name
  repository = "https://traefik.github.io/charts"
  chart      = "traefik"
  version    = var.chart_versions.traefik
  # Helm's wait timed out with the Traefik pod already ready, so scripts/deploy.sh
  # waits for the rollout instead.
  wait = false

  values = [yamlencode({
    ingressClass = {
      enabled        = true
      isDefaultClass = true
    }
    service = {
      type = "ClusterIP"
    }
    resources = {
      requests = { cpu = "50m", memory = "64Mi" }
      limits   = { cpu = "500m", memory = "128Mi" }
    }
  })]
}

resource "helm_release" "monitoring" {
  name       = "monitoring"
  namespace  = kubernetes_namespace_v1.monitoring.metadata[0].name
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = var.chart_versions.kube_prometheus_stack
  wait       = true
  timeout    = 900

  values = [yamlencode({
    # Not reachable or not useful on a single-node test cluster.
    kubeEtcd              = { enabled = false }
    kubeControllerManager = { enabled = false }
    kubeScheduler         = { enabled = false }
    kubeProxy             = { enabled = false }
    alertmanager          = { enabled = false }

    prometheus = {
      prometheusSpec = {
        retention = var.prometheus_retention
        # Discover monitors and rules from every namespace, not only this release.
        serviceMonitorSelectorNilUsesHelmValues = false
        ruleSelectorNilUsesHelmValues           = false
        podMonitorSelectorNilUsesHelmValues     = false
        probeSelectorNilUsesHelmValues          = false
        resources = {
          requests = { cpu = "100m", memory = "400Mi" }
          limits   = { memory = "1Gi" }
        }
      }
    }

    grafana = {
      adminPassword = var.grafana_admin_password
      persistence   = { enabled = false }
      resources = {
        requests = { cpu = "50m", memory = "256Mi" }
        limits   = { memory = "512Mi" }
      }
      "grafana.ini" = {
        "auth.anonymous" = {
          enabled  = var.grafana_anonymous_viewer
          org_role = "Viewer"
        }
      }
    }
  })]
}
