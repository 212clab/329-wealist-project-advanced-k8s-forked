# =============================================================================
# OIDC/IAM Configuration - Outputs
# =============================================================================

output "github_actions_role_arn" {
  description = "ARN of the IAM role for GitHub Actions (set as AWS_ROLE_ARN in GitHub Secrets)"
  value       = module.github_oidc.role_arn
}

output "oidc_provider_arn" {
  description = "ARN of the GitHub OIDC Provider"
  value       = module.github_oidc.oidc_provider_arn
}

output "role_name" {
  description = "Name of the IAM role"
  value       = module.github_oidc.role_name
}
