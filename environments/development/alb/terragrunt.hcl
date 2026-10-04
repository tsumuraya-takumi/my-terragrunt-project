locals {
  env = yamldecode(file(find_in_parent_folders("environment.yaml")))
}

include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/alb"
}

dependency "network" {
  config_path = "../network"
}

inputs = {
  environment       = local.env.environment
  project_name      = include.root.locals.project_name
  vpc_id            = dependency.network.outputs.vpc_id
  public_subnet_ids = dependency.network.outputs.public_subnet_ids
}