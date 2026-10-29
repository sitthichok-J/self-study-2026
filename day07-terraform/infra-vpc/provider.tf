# root นี้ใช้เพื่อ "อ่าน terraform plan" ของ VPC เท่านั้น ไม่ต้อง apply ไม่ต้องมีบัญชี AWS
# ปิดการเช็ก credential และ metadata ไว้ plan จึงทำงานแบบ offline ได้
provider "aws" {
  region                      = var.region
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}
