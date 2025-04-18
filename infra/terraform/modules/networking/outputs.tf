output "vpc_id" {
  description = "The ID of the VPC"
  value       = aws_vpc.album_store_vpc.id
}

output "public_subnets" {
  description = "List of IDs of public subnets"
  value       = aws_subnet.public[*].id
}

output "private_subnets" {
  description = "List of IDs of private subnets"
  value       = aws_subnet.private[*].id
}

output "alb_sg_id" {
  description = "The ID of the ALB security group"
  value       = aws_security_group.alb_sg.id
}

output "ecs_sg_id" {
  description = "The ID of the ECS security group"
  value       = aws_security_group.ecs_sg.id
}

output "rds_sg_id" {
  description = "The ID of the RDS security group"
  value       = aws_security_group.rds_sg.id
}

output "alb_target_group_arn" {
  description = "The ARN of the ALB target group"
  value       = aws_lb_target_group.album_store_tg.arn
}

output "alb_dns_name" {
  description = "The DNS name of the ALB"
  value       = aws_lb.album_store_alb.dns_name
}

output "alb_zone_id" {
  description = "The canonical hosted zone ID of the ALB"
  value       = aws_lb.album_store_alb.zone_id
}

output "alb_arn" {
  description = "The ARN of the ALB"
  value       = aws_lb.album_store_alb.arn
}
