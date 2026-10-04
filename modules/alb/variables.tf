variable "environment" {
  type = string
}

variable "project_name" {
  type = string
}

variable "vpc_id" {
  type        = string
  description = "ALBを配置するVPC ID"
}

variable "public_subnet_ids" {
  type        = map(string)
  description = "ALBを配置するPublicサブネットID（AZ名をキーにしたmap）"
}