# =============================================================================
# OIDC/IAM Configuration - Variables
# =============================================================================

variable "aws_account_id" {
  description = "AWS Account ID"
  type        = string
}

variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "ap-northeast-2"
}

variable "github_org" {
  description = "GitHub Organization name"
  type        = string
  default     = "OrangesCloud"
}

variable "github_repo" {
  description = "GitHub Repository name"
  type        = string
  default     = "wealist-project-advanced-k8s"
}

variable "allowed_branches" {
  description = "List of branch patterns allowed to assume the role"
  type        = list(string)
  default = [
    "service-deploy-dev",
    "service-deploy-prod",
    "k8s-deploy-dev",
    "k8s-deploy-prod",
    "dev",
    "main"
  ]
}

variable "role_name" {
  description = "Name of the IAM role for GitHub Actions"
  type        = string
  default     = "wealist-github-actions-role"
}

variable "enable_s3_access" {
  description = "Enable S3 access for frontend deployment"
  type        = bool
  default     = true
}

variable "s3_bucket_names" {
  description = "List of S3 bucket names for frontend deployment"
  type        = list(string)
  default     = []
}

variable "enable_cloudfront_access" {
  description = "Enable CloudFront access for cache invalidation"
  type        = bool
  default     = true
}
