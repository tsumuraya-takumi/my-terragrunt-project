# ------------------------
# ALB
# ------------------------

resource "aws_alb" "main" {
  name               = "${var.project_name}-${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb_sg.id]
  subnets            = values(var.public_subnet_ids)

  tags = {
    Name = "${var.project_name}-${var.environment}-alb"
    Env  = var.environment
  }
}

# ------------------------
# Target Group
# ------------------------

resource "aws_lb_target_group" "tg-app" {
  name     = "${var.project_name}-${var.environment}-tg-app"
  port     = 80
  protocol = "HTTP"
  vpc_id   = var.vpc_id

  health_check {
    path                = "/"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-tg-app"
    Env  = var.environment
  }
}

# ------------------------
# Listener (HTTP)
# ------------------------

resource "aws_lb_listener" "ln-http" {
  load_balancer_arn = aws_alb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.tg-app.arn
  }
}

# ------------------------
# Security Group (ALB)
# ------------------------

resource "aws_security_group" "alb_sg" {
  name        = "${var.project_name}-${var.environment}-alb-sg"
  description = "Security group for ALB"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-alb-sg"
    Env  = var.environment
  }
}