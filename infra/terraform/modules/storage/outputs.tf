output "album_covers_bucket_name" {
  description = "Name of the S3 bucket for album covers"
  value       = aws_s3_bucket.album_covers.id
}

output "album_covers_bucket_arn" {
  description = "ARN of the S3 bucket for album covers"
  value       = aws_s3_bucket.album_covers.arn
}

output "album_covers_bucket_domain_name" {
  description = "Domain name of the S3 bucket for album covers"
  value       = aws_s3_bucket.album_covers.bucket_domain_name
}