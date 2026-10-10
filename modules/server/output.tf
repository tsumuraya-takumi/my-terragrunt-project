output "app_sg_id" {
  value = aws_security_group.app_sg.id
}

output "asg_name" {
  value = aws_autoscaling_group.app-asg.name
}