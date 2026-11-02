output "cluster_name" {
  value = module.eks.cluster_name
}

output "ecr_repository_url" {
  value = module.ecr.repository_url
}

output "dynamodb_table" {
  value = module.dynamodb.table_name
}

output "api_role_arn" {
  description = "Set it as the role-arn annotation in deploy/k8s/overlays/aws."
  value       = aws_iam_role.api.arn
}
