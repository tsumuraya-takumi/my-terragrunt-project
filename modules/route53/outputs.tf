output "zone_id" {
  value       = aws_route53_zone.hosted_zone.zone_id
  description = "ホステッドゾーンのID"
}

output "zone_name" {
  value       = aws_route53_zone.hosted_zone.name
  description = "ホステッドゾーンの名前"
}

output "name_servers" {
  value       = aws_route53_zone.hosted_zone.name_servers
  description = "お名前.comのネームサーバ設定に登録する値"
}