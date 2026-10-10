# ------------------------
# Public / Private Subnet
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