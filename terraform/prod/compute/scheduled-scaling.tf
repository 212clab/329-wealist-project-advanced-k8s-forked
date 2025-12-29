# =============================================================================
# Scheduled Scaling for EKS Node Groups
# =============================================================================
# 비용 절감을 위한 노드 그룹 자동 on/off 스케줄링
#
# 기본 스케줄 (KST 기준):
#   평일 (월-금):
#     - Scale Down: 새벽 01:00 (노드 종료)
#     - Scale Up:   오전 08:00 (노드 시작)
#   주말 (토-일):
#     - Scale Down: 새벽 03:00 (노드 종료)
#     - Scale Up:   오전 09:00 (노드 시작)
#
# 사용법:
#   scheduled_scaling_enabled = true   # 스케줄 활성화
#   scheduled_scaling_enabled = false  # 스케줄 비활성화 (24시간 운영)
#   weekend_enabled = true             # 주말도 스케줄 적용 (기본값)
# =============================================================================

# -----------------------------------------------------------------------------
# Data Source: Node Group ASG 이름 조회
# -----------------------------------------------------------------------------
data "aws_autoscaling_groups" "eks_node_groups" {
  count = var.scheduled_scaling_enabled ? 1 : 0

  filter {
    name   = "tag:eks:cluster-name"
    values = [module.eks.cluster_name]
  }

  filter {
    name   = "tag:eks:nodegroup-name"
    values = ["${local.name_prefix}-spot*"]
  }

  depends_on = [module.eks]
}

# =============================================================================
# 평일 스케줄 (월-금)
# =============================================================================

# -----------------------------------------------------------------------------
# Weekday Scale Down - 새벽 01:00 KST (16:00 UTC Sun-Thu)
# -----------------------------------------------------------------------------
resource "aws_autoscaling_schedule" "weekday_scale_down" {
  count = var.scheduled_scaling_enabled ? 1 : 0

  scheduled_action_name  = "${local.name_prefix}-weekday-scale-down"
  autoscaling_group_name = data.aws_autoscaling_groups.eks_node_groups[0].names[0]

  # 16:00 UTC (Sun-Thu) = 01:00 KST (Mon-Fri)
  recurrence = var.weekday_scale_down_schedule

  min_size         = 0
  max_size         = 0
  desired_capacity = 0

  depends_on = [module.eks]
}

# -----------------------------------------------------------------------------
# Weekday Scale Up - 오전 08:00 KST (23:00 UTC Sun-Thu)
# -----------------------------------------------------------------------------
resource "aws_autoscaling_schedule" "weekday_scale_up" {
  count = var.scheduled_scaling_enabled ? 1 : 0

  scheduled_action_name  = "${local.name_prefix}-weekday-scale-up"
  autoscaling_group_name = data.aws_autoscaling_groups.eks_node_groups[0].names[0]

  # 23:00 UTC (Sun-Thu) = 08:00 KST (Mon-Fri)
  recurrence = var.weekday_scale_up_schedule

  min_size         = var.spot_min_size
  max_size         = var.spot_max_size
  desired_capacity = var.spot_desired_size

  depends_on = [module.eks]
}

# =============================================================================
# 주말 스케줄 (토-일)
# =============================================================================

# -----------------------------------------------------------------------------
# Weekend Scale Down - 새벽 03:00 KST (18:00 UTC Fri-Sat)
# -----------------------------------------------------------------------------
resource "aws_autoscaling_schedule" "weekend_scale_down" {
  count = var.scheduled_scaling_enabled && var.weekend_enabled ? 1 : 0

  scheduled_action_name  = "${local.name_prefix}-weekend-scale-down"
  autoscaling_group_name = data.aws_autoscaling_groups.eks_node_groups[0].names[0]

  # 18:00 UTC (Fri-Sat) = 03:00 KST (Sat-Sun)
  recurrence = var.weekend_scale_down_schedule

  min_size         = 0
  max_size         = 0
  desired_capacity = 0

  depends_on = [module.eks]
}

# -----------------------------------------------------------------------------
# Weekend Scale Up - 오전 09:00 KST (00:00 UTC Sat-Sun)
# -----------------------------------------------------------------------------
resource "aws_autoscaling_schedule" "weekend_scale_up" {
  count = var.scheduled_scaling_enabled && var.weekend_enabled ? 1 : 0

  scheduled_action_name  = "${local.name_prefix}-weekend-scale-up"
  autoscaling_group_name = data.aws_autoscaling_groups.eks_node_groups[0].names[0]

  # 00:00 UTC (Sat-Sun) = 09:00 KST (Sat-Sun)
  recurrence = var.weekend_scale_up_schedule

  min_size         = var.spot_min_size
  max_size         = var.spot_max_size
  desired_capacity = var.spot_desired_size

  depends_on = [module.eks]
}
