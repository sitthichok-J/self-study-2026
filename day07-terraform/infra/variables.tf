variable "region" {
  type        = string
  description = "AWS region"
  default     = "ap-southeast-1"
}

variable "name_prefix" {
  type        = string
  description = "prefix ของชื่อ resource ทั้งหมด"
  default     = "selfstudy"
}

variable "suffix" {
  type        = string
  description = "ตัวต่อท้ายชื่อ bucket ให้ไม่ซ้ำ (S3 ชื่อต้อง unique ทั้งโลก)"
  default     = "dev01"
}

variable "use_localstack" {
  type        = bool
  description = "true = ชี้ไปที่ LocalStack, false = ใช้ AWS จริง (ต้องมี credential ใน environment และ ระวังค่าใช้จ่าย)"
  default     = true
}

variable "localstack_endpoint" {
  type        = string
  description = "endpoint ของ LocalStack: ถ้ารัน terraform ใน container ในเครือข่ายเดียวกับ compose ใช้ค่า default, ถ้ารันบนเครื่องใช้ http://localhost:4566"
  default     = "http://localstack:4566"
}

variable "max_receive_count" {
  type        = number
  description = "จำนวนครั้งที่ consumer รับ message เดิมได้ก่อนย้ายเข้า DLQ"
  default     = 3
}
