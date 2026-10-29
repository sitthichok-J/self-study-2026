variable "region" {
  type        = string
  description = "AWS region"
  default     = "ap-southeast-1"
}

variable "name_prefix" {
  type        = string
  description = "prefix ของชื่อ resource"
  default     = "selfstudy"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR ของ VPC"
  default     = "10.20.0.0/16"
}

variable "azs" {
  type        = list(string)
  description = "Availability Zone ที่ใช้ (กำหนดเองเพื่อให้ plan ได้โดยไม่ต้องเรียก AWS)"
  default     = ["ap-southeast-1a", "ap-southeast-1b"]
}
