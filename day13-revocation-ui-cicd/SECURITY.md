# SECURITY.md — Day 13 (เพิกถอน token, rate limit ต่อ user, ตรวจเคลม)

ข้อมูลทั้งหมดเป็นข้อมูลสมมติ เอกสารนี้บอกว่าอะไรปิดแล้ว อะไรยังเปิดอยู่ ตามที่โค้ดและเทสต์ทำจริง ผลรันจริงอยู่ใน `notes/run-results.md`

## Threat

- ผู้ที่ได้ access token ของคนอื่นไป (จากเครื่อง, log, หน้าจอ) แล้วใช้ต่อหลังเจ้าตัว logout
- ผู้ที่ขโมย refresh token ได้ แล้วใช้ก่อนหรือหลังเจ้าตัว
- บัญชีเดียวที่ยิง API ถี่ผ่านหลาย IP (Kong นับต่อ IP จึงไม่เห็น)
- ผู้ใช้ธรรมดาที่พยายามตัดสินเคลมของตัวเอง
- ไม่อยู่ใน scope: ผู้โจมตีที่คุมเครื่องหรือ container host, supply chain ของ base image

## ปิดแล้ว (ต่อจาก Day 12: RS256 + kid, หมุน key, log ไม่มี query, CSRF, cookie, rate limit ต่อ IP)

- **logout เพิกถอน access token ทันที**: UI ส่ง access token ปัจจุบันไปกับ `/auth/logout` API ขึ้นบัญชี `rev:jti:<jti>` ใน valkey จนถึง exp + leeway ทุกคำขอหลังจากนั้นด้วย token เดิมได้ 401 (Kong ยังผ่านเพราะตรวจแค่ลายเซ็นกับ exp แต่ API ตัดทิ้ง)
- **refresh token ถูกใช้ซ้ำ = access token ของ user นั้นที่ออกก่อนหน้าตายหมด** (`rev:user:<user>` เทียบ `iat`) แล้วค่อย login ใหม่ได้
- **ไม่รับ token ที่ไม่มี `jti`** (ใส่บัญชีไม่ได้ก็ไม่ควรรับ)
- **ตรวจ blocklist ไม่ได้ = ปิดไว้ก่อน**: valkey ล่มได้ 503 + `Retry-After` ไม่ใช่ 401 (UI จะไม่เด้ง user ออก) และไม่ใช่ 200
- **valkey เปิด AOF + volume**: blocklist อยู่รอดตอน restart (ทดสอบจริงใน run-all)
- **ไม่เก็บ token ใน valkey**: เก็บแค่ `jti` กับเวลา และทุกรายการมี TTL หมดเอง
- **rate limit ต่อ user** (ค่าเริ่มต้น 100/นาที, `USER_RATE_LIMIT`): นับเฉพาะคำขอที่ผ่านการตรวจ token (token มั่วไม่เผาโควต้าของคนอื่น), 429 + `Retry-After`
- **ตัดสินเคลม**: เฉพาะ staff (API ตรวจสิทธิ์จาก token เอง ซ่อนปุ่มใน UI ไม่ใช่การป้องกัน), ตัดสินได้เฉพาะ PENDING, เงื่อนไขแข่งกัน (staff สองคน หรือ worker กับ staff) ใช้ `UPDATE ... WHERE status='PENDING'` ผู้แพ้ได้ 409, เหตุผลเป็นรหัสคงที่ ไม่รับข้อความอิสระ (กัน PII หลุดลงข้อมูล)
- **Kong Admin API / Kong Manager / Kafka UI ผูก 127.0.0.1 เท่านั้น** Kafka UI ตั้ง read-only และ static test ตรวจว่าทุกพอร์ตที่ publish เป็น loopback
- **CI**: สิทธิ์ `contents: read`, ไม่ใช้ `pull_request_target`, ไม่มี secret, action ทุกตัวระบุเวอร์ชัน, image ไม่ถูก push

## ยังไม่ปิด

- **valkey ตัวเดียว ไม่มี auth ไม่มี TLS** (อยู่ใน network ของ compose + ผูก 127.0.0.1) ใครเขียน valkey ได้ก็ล้าง blocklist ได้ ถ้า volume หายหรือ `docker compose down -v` รายการหายทั้งหมด (token ที่เพิกถอนไว้ใช้ได้อีกจน exp ≤10 นาที)
- **fail closed ทำให้ valkey ล่ม = API ใช้ไม่ได้** (ตั้งใจเลือก) ไม่มี replica ไม่มี failover
- **Kong ยังไม่รู้เรื่องการเพิกถอน** ตรวจแค่ลายเซ็นกับ exp ถ้ามีทางเรียก API โดยไม่ผ่านการตรวจของ API ชั้นหลัง token ที่ logout แล้วจะผ่าน (ตอนนี้ API เป็นด่านเดียวหลัง Kong)
- **rate limit ต่อ user เป็น fixed window**: ยิงติดกันตอนปลายหน้าต่างกับต้นหน้าต่างถัดไปได้เกินเพดานชั่วขณะ (สูงสุดราว 2 เท่า)
- **ไม่มี "logout ทุกเครื่อง" ให้ user กดเอง** มีแค่กรณีตรวจพบการใช้ refresh token ซ้ำ ไม่มีการเพิกถอนตอนเปลี่ยนรหัสผ่าน (ยังไม่มีฟังก์ชันเปลี่ยนรหัสผ่าน)
- **iat หน่วยวินาที**: token ที่ออกในวินาทีเดียวกับเหตุการณ์ถูกบล็อกด้วย (login ทันทีหลังเหตุการณ์ในวินาทีเดียวกันต้องรอข้ามวินาที)
- **Kong Manager โหมด DB-less เป็นหน้าอ่านอย่างเดียว** แต่ Admin API (:8001) ยังรับ `POST /config` ได้ ป้องกันด้วยการผูก loopback อย่างเดียว ใครบนเครื่องนี้เรียกได้
- **CI รันบน GitHub Actions แล้ว 5 รอบ สองรอบแรกพังจริงและแก้แล้ว**: README ขาดข้อความ stdlib, `trivy-action` ดึงไม่ได้ (เปลี่ยนไปรัน image Trivy ที่ pin เวอร์ชัน), govulncheck เจอช่องโหว่ stdlib ของ Go 1.26.0 (เปลี่ยนไปใช้ 1.26 ล่าสุด), image web มีแพ็กเกจ OS ที่แพตช์ได้ (เพิ่ม `apk upgrade`), gitleaks เจอ token ปลอมในเทสต์เพราะ path ใน CI ไม่ตรงกับ `.gitleaksignore` (แก้ที่ path) รอบที่ 5 ผ่านครบทุก job ยังไม่ได้รันผ่าน `act` และ CI ยังไม่เคยทดสอบกับ pull request
- **สแกนและ CI ไม่ครอบคลุม** base image ของ Kafka UI (`provectuslabs/kafka-ui`) และ Kong ที่ใช้แค่รันบนเครื่อง
- ข้อที่ยังค้างจาก Day 12: private key เป็นไฟล์ใน `keys/`, UI ผ่าน http://localhost:3000, cert self-signed, ไม่มีระบบจัดการ user/MFA/OIDC, audit log ถาวร
- **gitleaks และ Trivy fs เจอ secret ในโฟลเดอร์ที่รันบนเครื่อง** คือ `.env` (รหัส demo), `keys/k2.key`, `certs/server.key` ทั้งสามไฟล์ถูก git-ignore และไม่ถูกเอาเข้า repo (CI สแกนเฉพาะไฟล์ที่ checkout จึงไม่เจอ) ส่วนบน repo จริงหลัง push gitleaks รันใน CI แล้ว ผ่านหลังแก้ path ของ `.gitleaksignore` (รอบก่อนหน้าเจอ token ปลอมในเทสต์ 1 จุด)
- **ลอง UI ด้วยมือแล้ว (พิมพ์และคลิกเองผ่านเบราว์เซอร์) เฉพาะ**: login ops และ alice, ส่งเคลม 35,000 และ 30,000 บาท, staff กดอนุมัติ (ยืนยันสองครั้ง), alice ไม่เห็นปุ่ม, logout และดู Kong Manager, Kafka UI, Grafana, Prometheus, Jaeger, Redis Insight **ยังไม่ได้ลองด้วยมือ**: ปุ่มไม่อนุมัติ, ข้อความ 429 ของ rate limit ต่อ user, หน้า UI ตอน valkey ล่ม, เบราว์เซอร์อื่นนอกจาก Sidekick (ส่วนเหล่านี้ตรวจผ่าน API ใน run-all และเทสต์ Vue)
