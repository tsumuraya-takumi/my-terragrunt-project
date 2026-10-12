# ------------------------
# Auto Scaling Group
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# Target Groupが適切に設定されており、ヘルスチェックがELBに設定されているか
run "asg_is_registered_to_target_group" {
  command = plan

  # 変数で渡したTarget Groupに登録されていること
  assert {
    condition     = aws_autoscaling_group.app.target_group_arns == toset([var.target_group_arn])
    error_message = "ASGが指定したTarget Groupに登録されていません"
  }

  # ALBのヘルスチェックの結果でインスタンスの正常と異常を判定すること
  assert {
    condition     = aws_autoscaling_group.app.health_check_type == "ELB"
    error_message = "ASGのヘルスチェックタイプがELBになっていません"
  }
}

# Auto Scaling Group がプライベートサブネット上に配置されていること
run "asg_is_placed_in_private_subnets" {
  command = plan

  assert {
    condition     = aws_autoscaling_group.app.vpc_zone_identifier == toset(values(var.private_subnet_ids))
    error_message = "ASGがプライベートサブネット(1a, 1c)に配置されていません"
  }
}
