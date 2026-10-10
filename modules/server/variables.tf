variable "environment" {
  type = string
}

variable "project_name" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type        = map(string)
  description = "ASGを配置するプライベートサブネットID"
}

variable "alb_sg_id" {
  type        = string
  description = "ALBのSGのID"
}

variable "instance_type" {
  type = string
}

variable "target_group_arn" {
  type = string
}

variable "min_size" {
  type = number
}

variable "desired_capacity" {
  type = number
}

variable "max_size" {
  type = number
}