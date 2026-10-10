locals {
  name_prefix = "${var.project_name}-${var.environment}"
}

# ------------------------
# VPC
# ------------------------

resource "aws_vpc" "vpc" {
  cidr_block                       = var.vpc_cidr
  instance_tenancy                 = "default"
  enable_dns_support               = true
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = false

  tags = {
    Name = "${local.name_prefix}-vpc"
  }
}

# ------------------------
# Public / Private Subnet
# ------------------------

# Public Cidr (10.0.0.0/24、10.0.1.0/24)
resource "aws_subnet" "public_subnet" {
  for_each = toset(var.azs)

  vpc_id                  = aws_vpc.vpc.id
  availability_zone       = each.value
  cidr_block              = cidrsubnet(aws_vpc.vpc.cidr_block, 8, index(var.azs, each.value))
  map_public_ip_on_launch = true

  tags = {
    Name = "${local.name_prefix}-public-${substr(each.value, -2, 2)}"
  }
}

# Private Cidr (10.0.10.0/24、10.0.11.0/24)
resource "aws_subnet" "private_subnet" {
  for_each = toset(var.azs)

  vpc_id            = aws_vpc.vpc.id
  availability_zone = each.value
  cidr_block        = cidrsubnet(aws_vpc.vpc.cidr_block, 8, index(var.azs, each.value) + 10)

  tags = {
    Name = "${local.name_prefix}-private-${substr(each.value, -2, 2)}"
  }
}


# ------------------------
# Internet Gateway
# ------------------------

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${local.name_prefix}-igw"
  }
}


# ------------------------
# Route Table
# ------------------------

resource "aws_route_table" "public_rt" {
  for_each = toset(var.azs)

  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${local.name_prefix}-public-rt-${substr(each.value, -2, 2)}"
  }
}

resource "aws_route" "public_rt_default" {
  for_each = toset(var.azs)

  route_table_id         = aws_route_table.public_rt[each.value].id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

resource "aws_route_table_association" "public_rt_assoc" {
  for_each = toset(var.azs)

  route_table_id = aws_route_table.public_rt[each.value].id
  subnet_id      = aws_subnet.public_subnet[each.value].id
}

# ------------------------
# NAT Gateway
# ------------------------

resource "aws_eip" "nat" {
  for_each = toset(var.azs)

  domain = "vpc"

  tags = {
    Name = "${local.name_prefix}-nat-eip-${substr(each.value, -2, 2)}"
  }
}

resource "aws_nat_gateway" "nat" {
  for_each = toset(var.azs)

  allocation_id = aws_eip.nat[each.value].id
  subnet_id     = aws_subnet.public_subnet[each.value].id

  tags = {
    Name = "${local.name_prefix}-nat-${substr(each.value, -2, 2)}"
  }

  depends_on = [aws_internet_gateway.igw]
}

# ------------------------
# Private Route Table
# ------------------------

resource "aws_route_table" "private_rt" {
  for_each = toset(var.azs)

  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${local.name_prefix}-private-rt-${substr(each.value, -2, 2)}"
  }
}

resource "aws_route" "private_rt_default" {
  for_each = toset(var.azs)

  route_table_id         = aws_route_table.private_rt[each.value].id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.nat[each.value].id
}

resource "aws_route_table_association" "private_rt_assoc" {
  for_each = toset(var.azs)

  route_table_id = aws_route_table.private_rt[each.value].id
  subnet_id      = aws_subnet.private_subnet[each.value].id
}
