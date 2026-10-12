# ------------------------
# Target Group
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# ヘルスチェックのパス、ポート、プロトコルはhttpdと一致していること
run "health_check_matches_httpd" {
  command = plan

  # 転送先(ターゲット)が80番(HTTP)であること
  assert {
    condition     = aws_lb_target_group.app.port == 80 && aws_lb_target_group.app.protocol == "HTTP"
    error_message = "ターゲットグループが80番(HTTP)になっていません"
  }

  # ヘルスチェックが httpd の「/」に対して HTTP で行われ、200 を正常とみなすこと
  assert {
    condition = (
      aws_lb_target_group.app.health_check[0].path == "/" &&
      aws_lb_target_group.app.health_check[0].port == "traffic-port" &&
      aws_lb_target_group.app.health_check[0].protocol == "HTTP" &&
      aws_lb_target_group.app.health_check[0].matcher == "200"
    )
    error_message = "ヘルスチェックの設定(/、traffic-port、HTTP、200)がhttpdと一致していません"
  }
}
