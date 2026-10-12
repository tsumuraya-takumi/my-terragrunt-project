# ------------------------
# Load Balancer
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# ロードバランサーがapplicationタイプ(ALB)が設定されていること
run "alb_is_internet_facing" {
  command = plan

  assert {
    condition     = aws_lb.main.load_balancer_type == "application"
    error_message = "ロードバランサーがALBではありません"
  }
}


# ALBがパブリックサブネット上に配置されていること
run "alb_is_placed_in_public_subnets" {
  command = plan

  assert {
    condition     = aws_lb.main.subnets == toset(values(var.public_subnet_ids))
    error_message = "パブリックサブネット上にALBがありません"
  }
}

