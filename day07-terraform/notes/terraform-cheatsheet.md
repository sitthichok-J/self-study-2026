# Terraform ใช้บ่อย

## วงจร
```
terraform init          # โหลด provider/module
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform apply tfplan
terraform output
terraform destroy
```
`plan -detailed-exitcode` ให้ exit code 0 = ไม่มีอะไรเปลี่ยน, 2 = มีการเปลี่ยนแปลง ใช้ใน CI ได้

## ชนิดของบล็อก
- `provider` ตั้งค่าการเชื่อมต่อ, `resource` สร้างของ, `data` อ่านของที่มีอยู่แล้ว
- `variable` รับค่าเข้า, `output` ส่งค่าออก, `locals` ตัวแปรภายใน
- `module` เรียกโฟลเดอร์อื่นเป็นหน่วยนำกลับมาใช้
- `count` / `for_each` สร้างหลายตัว (for_each ปลอดภัยกว่าเวลาลบรายการกลางลำดับ)

## state
- state คือไฟล์ที่ Terraform จำว่าสร้างอะไรไปแล้ว มี ID และบางครั้งมีค่า sensitive **ห้าม commit** (`.gitignore` มี `*.tfstate`)
- ทำงานเป็นทีมต้องใช้ remote state (เช่น S3 + lock) ไม่งั้นสองคน apply ชนกัน
  ตัวอย่าง (ยังไม่ได้เปิดใช้ในโค้ดวันนี้):
```hcl
terraform {
  backend "s3" {
    bucket       = "my-tf-state-bucket"
    key          = "day07/terraform.tfstate"
    region       = "ap-southeast-1"
    use_lockfile = true
  }
}
```
  (ตัวเลือก locking ของ backend s3 เปลี่ยนตามเวอร์ชัน ให้ตรวจเอกสารของเวอร์ชันที่ใช้)

## แยก environment
- โฟลเดอร์แยก (`envs/dev`, `envs/prod`) อ่านง่ายและ state แยกชัด หรือ workspace (state เดียวกัน backend, ชื่อ workspace ต่างกัน) เหมาะกับโครงสร้างเหมือนกันทุกตัว

## secret
- ห้ามใส่ใน `.tf` หรือ `.tfvars` ที่ commit ส่งผ่าน environment (`TF_VAR_xxx`) หรืออ่านจาก secret manager และจำไว้ว่าค่าที่ผ่าน Terraform อาจไปอยู่ใน state
