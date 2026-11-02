mock_provider "kubernetes" {}
mock_provider "helm" {}

run "installs_the_platform" {
  command = plan

  assert {
    condition     = module.platform.monitoring_namespace == "monitoring"
    error_message = "NetworkPolicy and the dashboard expect the monitoring namespace."
  }

  assert {
    condition     = module.platform.ingress_namespace == "ingress"
    error_message = "NetworkPolicy expects the ingress namespace."
  }
}
