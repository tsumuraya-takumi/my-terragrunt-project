# ------------------------
# Security Group
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# ALBのSGが、80ポートを許可しておりその他のポートは許可されていないこと
run "alb_sg_allows_http_from_internet" {
  command = plan

  # 受信ルールが1つだけであること（80番以外が開いていない）
  assert {
    condition     = length(aws_security_group.alb.ingress) == 1
    error_message = "ALBのSGに80番以外の受信ルールがあります"
  }

  # そのルールが80番(TCP)をインターネットから許可していること
  assert {
    condition = alltrue([
      for rule in aws_security_group.alb.ingress :
      rule.from_port == 80 &&
      rule.to_port == 80 &&
      rule.protocol == "tcp" &&
      rule.cidr_blocks == tolist(["0.0.0.0/0"])
    ])
    error_message = "ALBのSGは80ポートの許可がされてません"
  }
}