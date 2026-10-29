# credential ไม่ได้เขียนไว้ในไฟล์ ส่งผ่าน environment (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY)
# กับ LocalStack ใช้ค่าอะไรก็ได้ เช่น test
provider "aws" {
  region = var.region

  skip_credentials_validation = var.use_localstack
  skip_metadata_api_check     = var.use_localstack
  skip_requesting_account_id  = var.use_localstack
  s3_use_path_style           = var.use_localstack

  dynamic "endpoints" {
    for_each = var.use_localstack ? [1] : []
    content {
      s3  = var.localstack_endpoint
      sqs = var.localstack_endpoint
      iam = var.localstack_endpoint
      sts = var.localstack_endpoint
      ec2 = var.localstack_endpoint
    }
  }
}
