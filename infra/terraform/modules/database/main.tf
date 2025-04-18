// Create a subnet group for RDS
resource "aws_db_subnet_group" "album_store_db_subnet_group" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

// Create an RDS instance for MySQL
resource "aws_db_instance" "album_store_mysql" {
  identifier             = "${var.project_name}-mysql"
  engine                 = "mysql"
  engine_version         = "8.0"
  instance_class         = var.db_instance_class
  allocated_storage      = var.db_allocated_storage
  db_name                = var.db_name
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.album_store_db_subnet_group.name
  vpc_security_group_ids = [var.rds_security_group_id]
  skip_final_snapshot    = true
  publicly_accessible    = false
  deletion_protection    = false

  tags = {
    Name = "${var.project_name}-mysql"
  }
}

// Create a DynamoDB table for album reviews
resource "aws_dynamodb_table" "album_reviews" {
  name         = var.dynamodb_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "album_id"

  attribute {
    name = "album_id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "${var.project_name}-album-reviews"
  }
}
