# ------------------------
# Launch Template
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# user_data の一行目が #!/bin/bash から始まっていることを確認
# （先頭にスペースがあると cloud-init がスクリプトとして実行しない）
run "user_data_starts_with_shebang" {
  command = plan

  assert {
    condition     = startswith(base64decode(aws_launch_template.app.user_data), "#!/bin/bash\n")
    error_message = "user_data の一行目が #!/bin/bash で始まっていません"
  }
}

# user_data でhttpdサービスのインストールと自動起動が設定されていること
run "user_data_installs_and_enables_httpd" {
  command = plan

  assert {
    condition     = strcontains(base64decode(aws_launch_template.app.user_data), "dnf install -y httpd")
    error_message = "user_data に httpd のインストール（dnf install -y httpd）がありません"
  }

  assert {
    condition     = strcontains(base64decode(aws_launch_template.app.user_data), "systemctl enable --now httpd")
    error_message = "user_data に httpd の自動起動設定（systemctl enable --now httpd）がありません"
  }
}

# Launch Templateにapp用のSGが設定されていること
# （未設定だとVPCのdefault SGが使われ、ALBからの通信が届かない）
run "launch_template_uses_app_sg" {
  command = apply

  # mock のランダムIDは ASG の入力チェック（lt- で始まること）に通らないため固定する
  override_resource {
    target = aws_launch_template.app
    values = {
      id = "lt-00000000000000000"
    }
  }

  assert {
    condition     = aws_launch_template.app.vpc_security_group_ids == toset([aws_security_group.app.id])
    error_message = "Launch Templateにapp用のSGが設定されていません"
  }
}
