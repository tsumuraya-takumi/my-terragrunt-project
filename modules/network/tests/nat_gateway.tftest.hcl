# ------------------------
# NAT Gateway
# ------------------------

mock_provider "aws" {
  override_during = plan
}

# NAT GWがパブリックサブネット上に存在すること

run "nat_gw_is_placed_in_public_subnet" {
  command = apply

  assert {
    condition = alltrue([
      for az, nat in aws_nat_gateway.nat : nat.subnet_id == aws_subnet.public_subnet[az].id
    ])
    error_message = "パブリックサブネット上に配置されていないNAT GWがあります"
  }
}

# NAT GWに同じAZ用のEIPが割り当てられているか

run "nat_gw_uses_eip_of_same_az" {
  command = apply

  assert {
    condition = alltrue([
      for az, nat in aws_nat_gateway.nat : nat.allocation_id == aws_eip.nat[az].id
    ])
    error_message = "同じAZ用のEIPが割り当てられていないNAT GWがあります"
  }
}
