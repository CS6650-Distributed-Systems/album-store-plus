output "ecs_cluster_id" {
  description = "ID of the ECS cluster"
  value       = aws_ecs_cluster.album_store_cluster.id
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.album_store_cluster.name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.album_store_service.name
}

output "worker_service_name" {
  description = "Name of the worker ECS service"
  value       = aws_ecs_service.worker_service.name
}

output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.album_store_task.arn
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group for ECS logs"
  value       = aws_cloudwatch_log_group.ecs_logs.name
}
