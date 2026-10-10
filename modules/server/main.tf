# ------------------------
# Security Group (App)
# ------------------------

resource "aws_security_group" "app_sg" {
  name        = "${var.project_name}-${var.environment}-app-sg"
  description = "Security group for app instances"
  vpc_id = var.vpc_id

  ingress {
    description = "Allow HTTP for ALB"
    from_port = 80
    to_port = 80
    protocol = "tcp"
    security_groups = [var.alb_sg_id]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port = 0
    to_port = 0
    protocol = -1
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name  = "${var.project_name}-${var.environment}-app-sg"
    Env = var.environment
  }
}

# ------------------------
# Launch Template
# ------------------------

resource "aws_launch_template" "app-lt" {
  name_prefix = "${var.project_name}-${var.environment}-app-"
  image_id = "resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"
  instance_type = var.instance_type
  vpc_security_group_ids = [aws_security_group.app_sg.id]

  metadata_options {
    http_tokens = "required"
  }

  user_data = base64encode(<<EOF
#!/bin/bash
dnf install -y httpd
systemctl enable --now httpd
echo "<h1>$(hostname -f)</h1>" > /var/www/html/index.html
EOF
  )

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "${var.project_name}-${var.environment}-app-lt"
      Env = var.environment
    }
  }
}

# ------------------------
# Auto Scaling Group
# ------------------------

resource "aws_autoscaling_group" "app-asg" {
  name = "${var.project_name}-${var.environment}-app-asg"
  min_size = var.min_size
  desired_capacity = var.desired_capacity
  max_size = var.max_size
  vpc_zone_identifier = values(var.private_subnet_ids)
  target_group_arns = [var.target_group_arn]
  health_check_type = "ELB"
  health_check_grace_period = 300


  launch_template {
    id = aws_launch_template.app-lt.id
    version = "$Latest"
  }

   tag {
    key                 = "Name"
    value               = "${var.project_name}-${var.environment}-app"
    propagate_at_launch = true
  }

  tag {
    key                 = "Env"
    value               = var.environment
    propagate_at_launch = true
  }
}

