# ------------------------
# Public / Private Subnet
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# Public / Private サブネットのCIDRブロックの確認

run "public_subnet_has_correct_cidr" {
  command = plan

  assert {
    condition     = aws_subnet.public_subnet["ap-northeast-1a"].cidr_block == "10.0.0.0/24"
    error_message = "Public Subnet(1a)のCIDRブロックが10.0.0.0/24ではありません"
  }

  assert {
    condition     = aws_subnet.public_subnet["ap-northeast-1c"].cidr_block == "10.0.1.0/24"
    error_message = "Public Subnet(1c)のCIDRブロックが10.0.1.0/24ではありません"
  }
}

run "priate_subnet_has_correct_cidr" {
  command = plan

  assert {
    condition     = aws_subnet.private_subnet["ap-northeast-1a"].cidr_block == "10.0.10.0/24"
    error_message = "Private Subnet(1a)のCIDRブロックが10.0.10.0/24ではありません"
  }

  assert {
    condition     = aws_subnet.private_subnet["ap-northeast-1c"].cidr_block == "10.0.11.0/24"
    error_message = "Private Subnet(1c)のCIDRブロックが10.0.11.0/24ではありません"
  }
}

# Public / Private Subnet が明確に区別できているか

run "map_public_ip_settings_are_correct" {
  command = plan

  assert {
    condition     = aws_subnet.public_subnet["ap-northeast-1a"].map_public_ip_on_launch == true
    error_message = "Public Subnetでパブリックipの自動割り当てが有効になっていません"
  }

  assert {
    condition     = aws_subnet.private_subnet["ap-northeast-1a"].map_public_ip_on_launch != true
    error_message = "Private Subnetでパブリックipの自動割り当てが有効になってしまっています"
  }
}

# Public / Private Subnet タグの確認
run "subnet_tags_are_correct" {
  command = plan

  assert {
    condition     = aws_subnet.public_subnet["ap-northeast-1a"].tags.Name == "my-project-development-public-1a"
    error_message = "Public SubnetのNameタグが想定と違います"
  }

  assert {
    condition     = aws_subnet.private_subnet["ap-northeast-1a"].tags.Name == "my-project-development-private-1a"
    error_message = "Private SubnetのNameタグが想定と違います"
  }
}