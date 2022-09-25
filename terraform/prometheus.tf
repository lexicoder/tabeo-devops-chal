resource "helm_release" "prometheus" {
  name       = "prometheus"
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "prometheus-community"
  namespace  = kubernetes_namespace_v1.monitoring.id
}
