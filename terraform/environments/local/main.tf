# Provisions the platform on an existing minikube cluster. scripts/local-up.sh and
# the CI pipeline create the cluster first. This environment costs nothing.

terraform {
  required_version = ">= 1.9"

  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.3"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.4"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.9"
    }
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kube_context
}

provider "helm" {
  kubernetes = {
    config_path    = var.kubeconfig_path
    config_context = var.kube_context
  }
}

# Generated here so no password lives in the repository. The state file holds it,
# and *.tfstate is ignored by git.
resource "random_password" "grafana_admin" {
  length  = 24
  special = false
}

module "platform" {
  source = "../../modules/platform"

  grafana_admin_password   = random_password.grafana_admin.result
  grafana_anonymous_viewer = true
  prometheus_retention     = "6h"
}
