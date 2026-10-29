output "vpc_id" {
  description = "id ของ VPC"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "subnet ที่เปิดออกอินเทอร์เน็ต"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "subnet ปิด สำหรับ app และ db"
  value       = aws_subnet.private[*].id
}
