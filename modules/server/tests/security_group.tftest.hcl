# ------------------------
# Security Group (App)
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# App SGへのHTTP(80)は、ALB SGからのみ許可されているか
# （インスタンスがインターネットに直接公開されていないこと）

run "app_sg_allows_http_only_from_alb_sg" {
  command = plan

  # 受信ルールが1つだけであること（意図しないルールが追加されていない）
  assert {
    condition     = length(aws_security_group.app.ingress) == 1
    error_message = "App SGの受信ルールが一つではありません"
  }

  # そのルールが80番(TCP)をALB SGからのみ許可していること
  assert {
    condition = alltrue([
      for rule in aws_security_group.app.ingress :
      rule.from_port == 80 &&
      rule.to_port == 80 &&
      rule.protocol == "tcp" &&
      rule.security_groups == toset([var.alb_sg_id])
    ])
    error_message = "App SGの受信ルールが80ポートをALBのSGからなみ許可されていません"
  }
}
