locals {
  params = yamldecode(file("${get_terragrunt_dir()}/server.yaml"))
}

include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/server"
}

dependency "network" {
  config_path = "../network"

  mock_outputs_allowed_terraform_commands = ["init", "validate", "plan"]
  mock_outputs = {
    vpc_id             = "vpc-00000000000000000"
    private_subnet_ids = { "ap-northeast-1a" = "subnet-00000000000000010", "ap-northeast-1c" = "subnet-00000000000000011" }
  }
}

dependency "alb" {
  config_path = "../alb"

  mock_outputs_allowed_terraform_commands = ["init", "validate", "plan"]
  mock_outputs = {
    alb_sg_id        = "sg-00000000000000000"
    target_group_arn = "arn:aws:elasticloadbalancing:ap-northeast-1:000000000000:targetgroup/mock/0000000000000000"
  }
}

inputs = {
  environment        = include.root.locals.environment
  project_name       = include.root.locals.project_name
  vpc_id             = dependency.network.outputs.vpc_id
  private_subnet_ids = dependency.network.outputs.private_subnet_ids
  alb_sg_id          = dependency.alb.outputs.alb_sg_id
  target_group_arn   = dependency.alb.outputs.target_group_arn
  instance_type      = local.params.instance_type
  min_size           = local.params.min_size
  desired_capacity   = local.params.desired_capacity
  max_size           = local.params.max_size
}
