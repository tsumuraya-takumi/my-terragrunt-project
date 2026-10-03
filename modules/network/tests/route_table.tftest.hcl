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
    condition     = aws_route_table.public_rt["ap-northeast-1a"].tags.Name == "my-terragrunt-project-development-public-rt-1a"
    error_message = "パブリックルートテーブル1aのNameタグが想定と違います"
  }
}

run "route_table_1c_tags_are_correct" {
  command = plan

  assert {
    condition     = aws_route_table.public_rt["ap-northeast-1c"].tags.Name == "my-terragrunt-project-development-public-rt-1c"
    error_message = "パブリックルートテーブル1cのNameタグが想定と違います"
  }
}

# IGWがVPCに正しく紐づいているか

run "internet_gw_is_attached_to_correct_vpc" {
  command = apply

  assert {
    condition     = aws_internet_gateway.igw.vpc_id != null
    error_message = "Internet GatewayがVPCに紐づいていません"
  }
}

# Route Tableのデフォルトルートが正しくIGWを向いているか

run "public_route_points_to_igw" {
  command = plan

  assert {
    condition     = aws_route.public_rt_default["ap-northeast-1a"].destination_cidr_block == "0.0.0.0/0"
    error_message = "Public Route Tableのデフォルトルートが0.0.0.0/0になっていません"
  }
}