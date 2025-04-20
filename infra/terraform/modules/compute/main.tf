// Create an ECS cluster
resource "aws_ecs_cluster" "album_store_cluster" {
  name = "${var.project_name}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name = "${var.project_name}-cluster"
  }
}

// Reference the existing ECS task execution role
data "aws_iam_role" "ecs_service_role" {
  name = var.iam_role_name
}

// Create a CloudWatch log group for ECS logs
resource "aws_cloudwatch_log_group" "ecs_logs" {
  name              = "/ecs/${var.project_name}"
  retention_in_days = 30

  tags = {
    Name = "${var.project_name}-ecs-logs"
  }
}

// Create a CloudWatch log group for Worker logs
resource "aws_cloudwatch_log_group" "worker_logs" {
  name              = "/ecs/${var.project_name}-worker"
  retention_in_days = 30

  tags = {
    Name = "${var.project_name}-worker-logs"
  }
}

// Create an ECS task definition
resource "aws_ecs_task_definition" "album_store_task" {
  family                   = "${var.project_name}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.fargate_cpu
  memory                   = var.fargate_memory
  execution_role_arn       = data.aws_iam_role.ecs_service_role.arn
  task_role_arn            = data.aws_iam_role.ecs_service_role.arn

  container_definitions = jsonencode([
    {
      name      = "${var.project_name}-container"
      image     = "${var.container_image}:latest"
      essential = true

      portMappings = [
        {
          containerPort = var.app_port
          hostPort      = var.app_port
          protocol      = "tcp"
        }
      ]

      environment = [
        { name = "ENVIRONMENT", value = var.environment },
        { name = "SERVER_PORT", value = tostring(var.app_port) },

        # MySQL config
        { name = "MYSQL_HOST", value = var.rds_endpoint },
        { name = "MYSQL_PORT", value = var.rds_port },
        { name = "MYSQL_USERNAME", value = var.rds_username },
        { name = "MYSQL_PASSWORD", value = var.rds_password },
        { name = "MYSQL_DATABASE", value = var.rds_database_name },

        # DynamoDB config
        { name = "DYNAMODB_TABLE_NAME", value = var.dynamodb_table_name },

        # S3 config
        { name = "S3_IMAGES_BUCKET", value = var.s3_bucket_name },

        # SNS config
        { name = "SNS_TOPIC_ARN", value = var.sns_topic_arn },

        # SQS config
        { name = "SQS_QUEUE_URL", value = var.sqs_queue_url },

        # Lambda config
        { name = "LAMBDA_FUNCTION_NAME", value = var.lambda_function_name },

        # AWS Region
        { name = "AWS_REGION", value = var.aws_region },

        # Feature flags
        { name = "FEATURE_USE_LOCAL_IMAGE_PROCESSING", value = "false" },
        { name = "FEATURE_USE_DYNAMODB_FOR_REVIEWS", value = "true" },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs_logs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-task"
  }
}

// Create an ECS task definition for the worker
resource "aws_ecs_task_definition" "worker_task" {
  family                   = "${var.project_name}-worker-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.worker_fargate_cpu
  memory                   = var.worker_fargate_memory
  execution_role_arn       = data.aws_iam_role.ecs_service_role.arn
  task_role_arn            = data.aws_iam_role.ecs_service_role.arn

  container_definitions = jsonencode([
    {
      name      = "${var.project_name}-worker-container"
      image     = "${var.worker_container_image}:latest"
      essential = true

      environment = [
        { name = "ENVIRONMENT", value = var.environment },
        # DynamoDB config
        { name = "DYNAMODB_TABLE_NAME", value = var.dynamodb_table_name },
        # SQS config
        { name = "SQS_QUEUE_URL", value = var.sqs_queue_url },
        # AWS Region
        { name = "AWS_REGION", value = var.aws_region }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.worker_logs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-worker-task"
  }
}

// Create an ECS service
resource "aws_ecs_service" "album_store_service" {
  name            = "${var.project_name}-service"
  cluster         = aws_ecs_cluster.album_store_cluster.id
  task_definition = aws_ecs_task_definition.album_store_task.arn
  desired_count   = var.app_count
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [var.ecs_security_group_id]
    subnets          = var.private_subnet_ids
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.alb_target_group_arn
    container_name   = "${var.project_name}-container"
    container_port   = var.app_port
  }

  tags = {
    Name = "${var.project_name}-service"
  }

  lifecycle {
    ignore_changes = [desired_count]
  }
}

// Create an ECS service for the worker
resource "aws_ecs_service" "worker_service" {
  name            = "${var.project_name}-worker-service"
  cluster         = aws_ecs_cluster.album_store_cluster.id
  task_definition = aws_ecs_task_definition.worker_task.arn
  desired_count   = var.worker_count
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [var.ecs_security_group_id]
    subnets          = var.private_subnet_ids
    assign_public_ip = false
  }

  tags = {
    Name = "${var.project_name}-worker-service"
  }
}
