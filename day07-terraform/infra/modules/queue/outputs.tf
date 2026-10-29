output "arn" {
  description = "ARN ของ queue หลัก"
  value       = aws_sqs_queue.this.arn
}

output "url" {
  description = "URL ของ queue หลัก"
  value       = aws_sqs_queue.this.url
}

output "dlq_arn" {
  description = "ARN ของ dead-letter queue"
  value       = aws_sqs_queue.dlq.arn
}
