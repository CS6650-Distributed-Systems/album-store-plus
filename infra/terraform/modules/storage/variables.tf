variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "album_covers_bucket_name" {
  description = "Name of the S3 bucket for album covers"
  type        = string
}

variable "app_repository_name" {
  description = "Name of the ECR repository for the main application"
  type        = string
  default     = "album-store-plus"
}

variable "worker_repository_name" {
  description = "Name of the ECR repository for the worker"
  type        = string
  default     = "album-store-worker"
}
