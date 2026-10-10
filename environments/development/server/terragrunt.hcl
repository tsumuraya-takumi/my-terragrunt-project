locals {
  env    = yamldecode(file(find_in_parent_folders("environment.yaml")))
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
}

dependency "alb" {
  config_path = "../alb"
}

inputs = {
  environment        = local.env.environment
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