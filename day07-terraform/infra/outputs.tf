output "bucket_name" {
  description = "ชื่อ S3 bucket"
  value       = aws_s3_bucket.docs.bucket
}

output "queue_url" {
  description = "URL ของ SQS queue หลัก"
  value       = module.claims_queue.url
}

output "dlq_arn" {
  description = "ARN ของ DLQ"
  value       = module.claims_queue.dlq_arn
}

output "consumer_role_arn" {
  description = "ARN ของ role ฝั่ง consumer"
  value       = aws_iam_role.consumer.arn
}
