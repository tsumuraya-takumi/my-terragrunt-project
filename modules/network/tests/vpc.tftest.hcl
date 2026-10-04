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

  assert {
    condition     = aws_vpc.vpc.tags.Env == "development"
    error_message = "環境がdevelopmentではありません"
  }
}
