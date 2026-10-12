# 共通の変数の定義
environment        = "development"
project_name       = "my-project"
vpc_id             = "vpc-00000000000000000"
private_subnet_ids = { "ap-northeast-1a" = "subnet-0000000000000000a", "ap-northeast-1c" = "subnet-0000000000000000c" }
alb_sg_id          = "sg-0000000000000000"
target_group_arn   = "arn:aws:elasticloadbalancing:ap-northeast-1:000000000000:targetgroup/test/0000000000000000"
instance_type      = "t3.micro"
min_size           = 2
desired_capacity   = 2
max_size           = 2