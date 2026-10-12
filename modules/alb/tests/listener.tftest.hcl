# ------------------------
# Listener
# ------------------------

mock_provider "aws" {
  # mock のランダム値は ARN の形式チェックに通らないため、ARN 形式の値を固定する
  mock_resource "aws_lb" {
    defaults = {
      arn = "arn:aws:elasticloadbalancing:ap-northeast-1:000000000000:loadbalancer/app/test/0000000000000000"
    }
  }

  mock_resource "aws_lb_target_group" {
    defaults = {
      arn = "arn:aws:elasticloadbalancing:ap-northeast-1:000000000000:targetgroup/test/0000000000000000"
    }
  }
}

# ALBからターゲットグループへHTTP(80ポート)通信ができること
run "http_listener_forwards_to_target_group" {
  command = apply

  assert {
    condition     = aws_lb_listener.http.port == 80 && aws_lb_listener.http.protocol == "HTTP"
    error_message = "リスナーが80番ポート(HTTP)で待ち受けていません"
  }

  assert {
    condition     = aws_lb_listener.http.load_balancer_arn == aws_lb.main.arn
    error_message = "リスナーがモジュールのALBに紐づいていません"
  }

  assert {
    condition = (
      aws_lb_listener.http.default_action[0].type == "forward" &&
      aws_lb_listener.http.default_action[0].target_group_arn == aws_lb_target_group.app.arn
    )
    error_message = "リスナーがモジュールのターゲットグループに転送していません"
  }
}