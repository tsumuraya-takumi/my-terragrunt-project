# ------------------------
# Route 53
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# ホストゾーンのドメイン名が変数で指定したものと一致するか確認
run "route53_zone_name_is_correct" {
  command = plan
  assert {
    condition     = aws_route53_zone.hosted_zone.name == var.domain_name
    error_message = "Route 53のホストゾーンのドメイン名が想定と違います"
  }
}

# Route 53 タグの確認
run "route53_tags_are_correct" {
  command = plan

  assert {
    condition     = aws_route53_zone.hosted_zone.tags.Name == "my-project-development-route53"
    error_message = "Route 53のNameタグが想定と違います"
  }
}