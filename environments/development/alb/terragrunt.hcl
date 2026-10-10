include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

terraform {
  source = "../../../modules/alb"
}

dependency "network" {
  config_path = "../network"

  mock_outputs_allowed_terraform_commands = ["init", "validate", "plan"]
  mock_outputs = {
    vpc_id            = "vpc-00000000000000000"
    public_subnet_ids = { "ap-northeast-1a" = "subnet-00000000000000000", "ap-northeast-1c" = "subnet-00000000000000001" }
  }
}

inputs = {
  environment       = include.root.locals.environment
  project_name      = include.root.locals.project_name
  vpc_id            = dependency.network.outputs.vpc_id
  public_subnet_ids = dependency.network.outputs.public_subnet_ids
}
