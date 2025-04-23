variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "create_dead_letter_queue" {
  description = "Whether to create a dead-letter queue for failed messages"
  type        = bool
  default     = true
}

variable "dead_letter_queue_arn" {
  description = "ARN of an existing dead-letter queue to use (if not creating a new one)"
  type        = string
  default     = ""
}
