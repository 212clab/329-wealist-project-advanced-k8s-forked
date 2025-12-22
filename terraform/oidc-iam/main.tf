# =============================================================================
# OIDC/IAM Configuration
# =============================================================================
# GitHub Actions → AWS 인증을 위한 OIDC Provider 및 IAM Role 설정
#
# 사용법:
#   1. terraform.tfvars.example을 terraform.tfvars로 복사
#   2. aws_account_id 설정
#   3. terraform init && terraform apply
#   4. 출력된 role_arn을 GitHub Secrets에 AWS_ROLE_ARN으로 등록

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# -----------------------------------------------------------------------------
# Provider Configuration
# -----------------------------------------------------------------------------
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "wealist"
      ManagedBy   = "terraform"
      Environment = "shared"
    }
  }
}

# -----------------------------------------------------------------------------
# GitHub OIDC Module
# -----------------------------------------------------------------------------
module "github_oidc" {
  source = "../modules/github-oidc"

  aws_account_id   = var.aws_account_id
  aws_region       = var.aws_region
  github_org       = var.github_org
  github_repo      = var.github_repo
  allowed_branches = var.allowed_branches
  role_name        = var.role_name

  enable_s3_access         = var.enable_s3_access
  s3_bucket_names          = var.s3_bucket_names
  enable_cloudfront_access = var.enable_cloudfront_access

  tags = {
    Purpose = "github-actions-oidc"
  }
}
