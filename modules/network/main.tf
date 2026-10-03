# ------------------------
# VPC
# ------------------------

resource "aws_vpc" "vpc" {
  cidr_block                       = "10.0.0.0/16"
  instance_tenancy                 = "default"
  enable_dns_support               = true
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = false

  tags = {
    Name = "${var.project_name}-${var.environment}-vpc"
    Env  = var.environment
  }
}

# ------------------------
# Public / Private Subnet
# ------------------------

# Public Cidr (10.0.0.0/24、10.0.1.0/24)
resource "aws_subnet" "public_subnet" {
  for_each = toset(var.public_subnet_azs)

  vpc_id                  = aws_vpc.vpc.id
  availability_zone       = each.value
  cidr_block              = cidrsubnet(aws_vpc.vpc.cidr_block, 8, index(var.public_subnet_azs, each.value))
  map_public_ip_on_launch = "true"

  tags = {
    Name = "${var.project_name}-${var.environment}-public-${substr(each.value, -2, 2)}"
    Env  = var.environment
  }
}

# Private Cidr (10.0.10.0/24、10.0.11.0/24)
resource "aws_subnet" "private_subnet" {
  for_each = toset(var.private_subnet_azs)

  vpc_id            = aws_vpc.vpc.id
  availability_zone = each.value
  cidr_block        = cidrsubnet(aws_vpc.vpc.cidr_block, 8, index(var.private_subnet_azs, each.value) + 10)

  tags = {
    Name = "${var.project_name}-${var.environment}-private-${substr(each.value, -2, 2)}"
    Env  = var.environment
  }
}


# ------------------------
# Internet Gateway
# ------------------------

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${var.project_name}-${var.environment}-igw"
    Env  = var.environment
  }
}


# ------------------------
# Route Table
# ------------------------

resource "aws_route_table" "public_rt" {
  for_each = toset(var.public_subnet_azs)

  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "${var.project_name}-${var.environment}-public-rt-${substr(each.value, -2, 2)}"
    Env  = var.environment
  }
}

resource "aws_route" "public_rt_default" {
  for_each = toset(var.public_subnet_azs)

  route_table_id         = aws_route_table.public_rt[each.value].id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

resource "aws_route_table_association" "public_rt_assoc" {
  for_each = toset(var.public_subnet_azs)

  route_table_id = aws_route_table.public_rt[each.value].id
  subnet_id      = aws_subnet.public_subnet[each.value].id
}