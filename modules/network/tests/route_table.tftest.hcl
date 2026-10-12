# ------------------------
# Route Table
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# Route table タグの確認

run "route_table_1a_tags_are_correct" {
  command = plan

  assert {
    condition     = aws_route_table.public_rt["ap-northeast-1a"].tags.Name == "my-project-development-public-rt-1a"
    error_message = "パブリックルートテーブル1aのNameタグが想定と違います"
  }
}

run "route_table_1c_tags_are_correct" {
  command = plan

  assert {
    condition     = aws_route_table.public_rt["ap-northeast-1c"].tags.Name == "my-project-development-public-rt-1c"
    error_message = "パブリックルートテーブル1cのNameタグが想定と違います"
  }
}

# Public用のRoute TableのデフォルトルートがIGWに向いているか

run "public_route_points_to_igw" {
  command = apply

  assert {
    condition = alltrue([
      for az, route in aws_route.public_rt_default :
      route.destination_cidr_block == "0.0.0.0/0" && route.gateway_id == aws_internet_gateway.igw.id
    ])
    error_message = "IGWを向いていないPublic Route Tableがあります"
  }
}

# Private用のRoute Tableのデフォルトルートが同じAZのNAT GWに向いているか

run "private_route_points_to_nat_gw" {
  command = apply

  assert {
    condition = alltrue([
      for az, route in aws_route.private_rt_default :
      route.destination_cidr_block == "0.0.0.0/0" && route.nat_gateway_id == aws_nat_gateway.nat[az].id
    ])
    error_message = "同じAZのNAT GWを向いていないPrivate Route Tableがあります"
  }
}

# 各サブネットが同じAZ・同じ種類（Public / Private）のRoute Tableに関連付いているか

run "public_subnets_are_associated_with_public_route_table" {
  command = apply

  assert {
    condition = alltrue([
      for az, assoc in aws_route_table_association.public_rt_assoc :
      assoc.subnet_id == aws_subnet.public_subnet[az].id && assoc.route_table_id == aws_route_table.public_rt[az].id
    ])
    error_message = "Public Route Tableに関連付いていないPublic Subnetがあります"
  }
}

run "private_subnets_are_associated_with_private_route_table" {
  command = apply

  assert {
    condition = alltrue([
      for az, assoc in aws_route_table_association.private_rt_assoc :
      assoc.subnet_id == aws_subnet.private_subnet[az].id && assoc.route_table_id == aws_route_table.private_rt[az].id
    ])
    error_message = "Private Route Tableに関連付いていないPrivate Subnetがあります"
  }
}
