locals {
  env    = yamldecode(file(find_in_parent_folders("environment.yaml")))
  params = yamldecode(file("${get_terragrunt_dir()}/network.yaml"))
}

include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/network"
}

inputs = {
  environment        = local.env.environment
  project_name       = include.root.locals.project_name
  public_subnet_azs  = local.params.public_subnet_azs
  private_subnet_azs = local.params.private_subnet_azs
}