variable "environment" {
  type = string
}

variable "project_name" {
  type = string
}

variable "vpc_cidr" {
  type        = string
  description = "VPCのCIDRブロック"
}

variable "azs" {
  type        = list(string)
  description = "サブネット・NAT Gatewayを配置するAZのリスト"
}
