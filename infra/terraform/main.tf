# terraform {
#   required_providers {
#     aws = {
#       source  = "hashicorp/aws"
#       version = "~> 5.0"
#     }
#   }
#
#   backend "s3" {
#     bucket         = "album-store-terraform-state"
#     key            = "terraform.tfstate"
#     region         = "us-west-2"
#     encrypt        = true
#     dynamodb_table = "album-store-terraform-lock"
#   }
# }

provider "aws" {
  region = var.aws_region
}

// Networking module
module "networking" {
  source = "./modules/networking"

  project_name       = var.project_name
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  app_port           = var.app_port
}

// Database module
module "database" {
  source = "./modules/database"

  project_name          = var.project_name
  private_subnet_ids    = module.networking.private_subnets
  rds_security_group_id = module.networking.rds_sg_id
  db_instance_class     = var.db_instance_class
  db_allocated_storage  = var.db_allocated_storage
  db_name               = var.db_name
  db_username           = var.db_username
  db_password           = var.db_password
  db_multi_az           = var.db_multi_az
  dynamodb_table_name   = var.dynamodb_table_name
}

// Messaging module
module "messaging" {
  source = "./modules/messaging"

  project_name            = var.project_name
  create_dead_letter_queue = true
}

// Serverless module (Lambda) - depends on the storage module
module "serverless" {
  source = "./modules/serverless"

  project_name       = var.project_name
  iam_role_name      = var.iam_role_name
  lambda_function_name = var.lambda_function_name
  lambda_handler     = var.lambda_handler
  lambda_runtime     = var.lambda_runtime
  lambda_zip_file    = var.lambda_zip_file
  lambda_memory_size = var.lambda_memory_size
  lambda_timeout     = var.lambda_timeout
  s3_bucket_name     = module.storage.album_covers_bucket_name
  max_image_width    = var.max_image_width
  max_image_height   = var.max_image_height
  image_quality      = var.image_quality
}

// Storage module (S3)
module "storage" {
  source = "./modules/storage"

  project_name            = var.project_name
  album_covers_bucket_name = var.album_covers_bucket_name
}

// Compute module (ECS)
module "compute" {
  source = "./modules/compute"

  project_name           = var.project_name
  environment            = var.environment
  aws_region             = var.aws_region
  iam_role_name          = var.iam_role_name
  app_port               = var.app_port
  app_count              = var.app_count
  worker_count           = var.worker_count
  fargate_cpu            = var.fargate_cpu
  worker_fargate_cpu     = var.worker_fargate_cpu
  fargate_memory         = var.fargate_memory
  worker_fargate_memory  = var.worker_fargate_memory
  container_image        = module.storage.app_repository_url
  worker_container_image = module.storage.worker_repository_url
  ecs_security_group_id  = module.networking.ecs_sg_id
  private_subnet_ids     = module.networking.private_subnets
  alb_target_group_arn   = module.networking.alb_target_group_arn
  s3_bucket_name         = module.storage.album_covers_bucket_name
  dynamodb_table_name    = module.database.dynamodb_table_name
  sns_topic_arn          = module.messaging.sns_topic_arn
  sqs_queue_url          = module.messaging.sqs_queue_url
  lambda_function_name   = module.serverless.lambda_function_name
  rds_endpoint           = module.database.rds_endpoint
  rds_port               = module.database.rds_port
  rds_username           = module.database.rds_username
  rds_password           = module.database.rdb_password
  rds_database_name      = module.database.rds_database_name
}
