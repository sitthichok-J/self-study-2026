# แบบฝึกหัด Day 7

ทำตามลำดับ ใช้ LocalStack ก่อน ถ้าจะใช้ AWS จริงให้ตั้ง billing alarm ก่อนและ destroy ทุกครั้ง

## 1. S3 + SQS (infra/)
1. `docker compose up -d` แล้วรอ localstack healthy
2. ในโฟลเดอร์ `infra`: `terraform init` → `plan -var-file=example.tfvars` อ่านว่าจะสร้างกี่ตัวและอะไรบ้าง ทายก่อนว่าได้กี่ตัวแล้วเทียบ
3. `apply` แล้วใช้ AWS CLI ชี้ endpoint LocalStack เช็ก: bucket มี public access block, queue มี redrive policy
4. ส่ง message เข้า queue แล้วรับด้วย `ReceiveMessage` 4 ครั้งโดยไม่ลบ ดูว่า message ย้ายเข้า DLQ หลังครั้งที่ 3 หรือไม่ (ต้องรอ visibility timeout หรือปรับค่าลงเพื่อทดลอง)
5. `plan` ซ้ำอีกรอบต้องไม่มีอะไรเปลี่ยน

## 2. อ่าน plan ของ VPC (infra-vpc/)
1. `terraform init && terraform plan` ไม่ต้อง apply
2. อ่านผลและตอบ: มี subnet กี่ตัว, route table กี่ตัว, ทำไม private ไม่มีเส้นทาง 0.0.0.0/0, security group ต่อกันเป็นโซ่อย่างไร
3. เปลี่ยน `azs` เป็น 3 ตัวแล้ว plan ใหม่ ดูว่า resource เพิ่มเท่าไร

## 3. IAM least privilege
1. อ่าน `infra/iam.tf` แล้วลองเพิ่ม action `sqs:PurgeQueue` ให้ consumer แล้วตอบว่าทำไมจึงไม่ควร
2. เขียน policy ใหม่ให้ service หนึ่งอ่าน object จาก bucket ได้อย่างเดียว

## 4. คิดเอง (เขียนสั้นๆ)
- ECS หรือ EKS สำหรับ Claim Tracker Lite เพราะอะไร
- ถ้า worker ล้มซ้ำๆ ข้อมูลไปอยู่ที่ไหนและใครต้องดูแล

## ก่อนเลิก
`terraform destroy` ทุก root ที่ apply แล้ว `docker compose down -v`
