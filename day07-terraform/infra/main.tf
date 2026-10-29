locals {
  tags = {
    project = "self-study-2026"
    day     = "07"
    managed = "terraform"
  }
}

# ---------------------------------------------------------------- S3
resource "aws_s3_bucket" "docs" {
  bucket = "${var.name_prefix}-docs-${var.suffix}"
  tags   = local.tags
}

resource "aws_s3_bucket_public_access_block" "docs" {
  bucket                  = aws_s3_bucket.docs.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "docs" {
  bucket = aws_s3_bucket.docs.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "docs" {
  bucket = aws_s3_bucket.docs.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# ---------------------------------------------------------------- SQS (module: queue + DLQ)
module "claims_queue" {
  source = "./modules/queue"

  name              = "${var.name_prefix}-claims"
  max_receive_count = var.max_receive_count
  tags              = local.tags
}
