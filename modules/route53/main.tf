locals {
  name_prefix = "${var.project_name}-${var.environment}"
}

# ------------------------
# Route 53
# ------------------------

resource "aws_route53_zone" "hosted_zone" {
  name = var.domain_name

  tags = {
    Name = "${local.name_prefix}-route53"
  }

  # 作り直すとネームサーバが変わり、お名前.com側の設定が必要になるため、削除しないようにする
  lifecycle {
    prevent_destroy = true
  }
}

