# ------------------------
# VPC
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# VPCのCidrブロックの確認

run "vpc_has_correct_cidr" {
  command = plan

  assert {
    condition     = aws_vpc.vpc.cidr_block == "10.0.0.0/16"
    error_message = "VPCのCIDRブロックが10.0.0.0/16になっていません"
  }
}

# VPCタグの確認
run "vpc_output_has_correct_tags" {
  command = plan

  assert {
    condition     = aws_vpc.vpc.tags.Name == "my-project-development-vpc"
    error_message = "VPCのNameタグが想定と違います"
  }
}

# VPCのDNS設定の確認（VPCエンドポイント利用時に必要）

run "vpc_dns_settings_are_enabled" {
  command = plan

  assert {
    condition     = aws_vpc.vpc.enable_dns_support == true
    error_message = "VPCのDNSサポートが有効になっていません"
  }

  assert {
    condition     = aws_vpc.vpc.enable_dns_hostnames == true
    error_message = "VPCのDNSホスト名が有効になっていません"
  }
}
