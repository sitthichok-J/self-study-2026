variable "name" {
  type        = string
  description = "ชื่อ queue หลัก (DLQ จะต่อท้ายด้วย -dlq)"
}

variable "max_receive_count" {
  type        = number
  description = "รับ message เดิมได้กี่ครั้งก่อนเข้า DLQ"
  default     = 3
}

variable "visibility_timeout_seconds" {
  type        = number
  description = "ต้องมากกว่าเวลาที่ consumer ใช้ประมวลผลหนึ่ง message"
  default     = 60
}

variable "tags" {
  type        = map(string)
  description = "tag ที่ติดทุก resource"
  default     = {}
}
