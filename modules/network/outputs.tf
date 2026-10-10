output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "vpc_tags" {
  value = aws_vpc.vpc.tags_all
}

output "public_subnet_ids" {
  value = { for az, subnet in aws_subnet.public_subnet : az => subnet.id }
}

output "private_subnet_ids" {
  value = { for az, subnet in aws_subnet.private_subnet : az => subnet.id }
}