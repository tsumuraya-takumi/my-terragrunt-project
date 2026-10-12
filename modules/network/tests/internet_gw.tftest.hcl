# ------------------------
# Internet Gateway
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# Internet Gateway タグの確認

run "internet_gw_tags_are_correct" {
  command = plan

  assert {
    condition     = aws_internet_gateway.igw.tags.Name == "my-project-development-igw"
    error_message = "Internet GatewayのNameタグが想定と違います"
  }
}

# IGWがこのモジュールのVPCに紐づいているか

run "internet_gw_is_attached_to_correct_vpc" {
  command = apply

  assert {
    condition     = aws_internet_gateway.igw.vpc_id == aws_vpc.vpc.id
    error_message = "Internet Gatewayがこのモジュールで作成したVPCに紐づいていません"
  }
}
