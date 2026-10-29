# AWS แนวคิดหลักสำหรับ backend (สรุปไว้ทบทวน)

## IAM
- **user** คือคนหรือโปรแกรมที่มี credential ระยะยาว (เลี่ยงให้ service ใช้) **role** คือสิทธิ์ที่ถูก assume ชั่วคราว ได้ credential ที่หมดอายุเอง จึงเป็นทางที่ควรใช้กับ service
- **policy** คือ JSON ที่บอก Effect / Action / Resource ทำแบบ least privilege: ระบุ action ที่ใช้จริงและระบุ resource เป็น ARN ไม่ใช้ `*`
- มี 2 ชนิดของ policy ที่ต้องแยก: identity policy (ติดกับ user/role บอกว่าทำอะไรได้) และ trust policy (ของ role บอกว่าใคร assume ได้ ในโค้ดคือ `assume_role_policy`)
- บน ECS ใช้ task role, บน EKS ใช้ IRSA ซึ่งแนวคิดเดียวกันคือผูกสิทธิ์กับ workload ไม่ใช่กับเครื่อง

## VPC
- VPC คือเครือข่ายส่วนตัวของเรา แบ่งเป็น subnet ต่อ AZ **public subnet** = route table มีเส้นทาง 0.0.0.0/0 ไปที่ Internet Gateway **private subnet** = ไม่มีเส้นทางนั้น
- ของที่ควรอยู่ใน private: app, database ส่วน ALB อยู่ public
- **Security Group** ทำงานระดับ instance/ENI, stateful (ตอบกลับอัตโนมัติ), มีแต่กฎ allow, อ้างอิง SG อื่นเป็นแหล่งที่มาได้ (เช่น db รับจาก app SG) **NACL** ทำงานระดับ subnet, stateless, มีทั้ง allow และ deny ปกติใช้ SG เป็นหลัก
- NAT Gateway ให้ private subnet ออกอินเทอร์เน็ตได้ แต่คิดเงินรายชั่วโมงและตามปริมาณข้อมูล โค้ดวันนี้จึงไม่สร้าง

## ECS vs EKS
| | ECS (Fargate) | EKS |
|---|---|---|
| ความซับซ้อน | ต่ำกว่า เป็นของ AWS ล้วน | สูงกว่า ต้องรู้ Kubernetes |
| ความยืดหยุ่น | เท่าที่ AWS ให้ | ใช้ ecosystem ของ Kubernetes ได้เต็ม (Helm, operator) |
| ค่าใช้จ่ายคงที่ | ไม่มีค่า control plane | control plane คิดเงินรายชั่วโมง |
| เหมาะกับ | ทีมเล็ก งานตรงไปตรงมา | ทีมที่ใช้ Kubernetes อยู่แล้ว (เช่น งานที่ลง k8s ตามไซต์ลูกค้า) |

ตรวจ pricing ล่าสุดของ AWS ก่อนตัดสินใจจริงเสมอ

## ข้อมูลและ messaging
- **RDS** = Postgres แบบ managed (backup, patch, multi-AZ) เปิด multi-AZ เมื่อต้องการ failover
- **S3** = object storage ปิด public access ทุก bucket เป็นค่าเริ่มต้น เปิด versioning และ encryption
- **SQS** = queue แบบ managed ไม่ต้องดูแล broker รองรับ DLQ ผ่าน redrive policy (`maxReceiveCount`) และ visibility timeout ต้องมากกว่าเวลาประมวลผล ไม่งั้น message ถูกส่งซ้ำ
- **SNS** = pub/sub หนึ่งข้อความไปหลาย subscriber ใช้คู่กับ SQS (fan-out) ได้
- เทียบกับ Kafka: SQS ง่ายกว่า ไม่มี offset/replay ลำดับไม่รับประกันนอกจาก FIFO queue ส่วน Kafka เก็บ log ให้ replay ได้และมี consumer group (AWS มี MSK เป็น Kafka แบบ managed)

## Well-Architected 6 ด้าน (ย่อ)
Operational excellence, Security, Reliability, Performance efficiency, Cost optimization, Sustainability
