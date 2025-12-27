# =============================================================================
# Production Compute - Variables
# =============================================================================

variable "aws_region" {
  description = "AWS Region"
  type        = string
  default     = "ap-northeast-2"
}

# =============================================================================
# EKS Cluster Configuration
# =============================================================================
variable "cluster_version" {
  description = "Kubernetes version for the EKS cluster"
  type        = string
  default     = "1.30"
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the EKS API endpoint"
  type        = list(string)
  default     = ["0.0.0.0/0"]  # 프로덕션에서는 제한 필요
}

# =============================================================================
# Node Group Configuration
# =============================================================================
variable "spot_instance_types" {
  description = "Instance types for Spot node group (다양한 타입으로 Spot 가용성 확보)"
  type        = list(string)
  default = [
    "t3.medium",   # 2 vCPU, 4GB RAM
    "t3a.medium",  # AMD 버전 (더 저렴)
    "t3.large",    # Fallback: 2 vCPU, 8GB RAM
    "t3a.large"    # Fallback AMD 버전
  ]
}

variable "spot_min_size" {
  description = "Minimum number of Spot nodes"
  type        = number
  default     = 2
}

variable "spot_max_size" {
  description = "Maximum number of Spot nodes"
  type        = number
  default     = 6
}

variable "spot_desired_size" {
  description = "Desired number of Spot nodes"
  type        = number
  default     = 3
}

variable "node_disk_size" {
  description = "Disk size in GB for worker nodes"
  type        = number
  default     = 50
}

# =============================================================================
# EKS Add-on Versions
# =============================================================================
variable "addon_versions" {
  description = "Versions for EKS managed add-ons"
  type = object({
    vpc_cni            = string
    coredns            = string
    kube_proxy         = string
    ebs_csi            = string
    pod_identity_agent = string
  })
  default = {
    vpc_cni            = "v1.18.5-eksbuild.1"
    coredns            = "v1.11.3-eksbuild.1"
    kube_proxy         = "v1.30.3-eksbuild.5"
    ebs_csi            = "v1.35.0-eksbuild.1"
    pod_identity_agent = "v1.3.2-eksbuild.2"
  }
}
