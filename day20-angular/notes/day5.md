# Day 20 — โน้ตวันที่ 5 (เทสต์ build CI)

## ทำอะไร
- ข้อ 1 ถึง 3 ของแผน (เทส ClaimCard/ClaimService, เทสฟอร์มและ validator, baseline + ทำให้พัง) ทำไปแล้วตั้งแต่ Day 16 ถึง 19 วันนี้ไล่ตรวจว่ายังครบ แล้วอุดช่องโหว่ที่ค้างจาก Day 19
- `app.config.spec.ts`: เปิดแอปด้วย providers ของ `appConfig` จริง นำทางไป `/claims/CLM-002` ผ่าน RouterTestingHarness แล้วเช็ค `detail.id()` ว่าเป็น `CLM-002` (ถ้าไม่มี `withComponentInputBinding()` input ไม่ถูกเติมและเทสต์ล้ม) และมีเทสต์ baseline นับ provider (3 ตัว)
- budget: initial เตือน 300kB พัง 350kB ลดจากค่าเริ่มต้นของ CLI เพื่อให้ bundle โตแล้วรู้ตัว
- `npm run test:ci` / `npm run ci` ให้ CI กับเครื่องรันคำสั่งเดียวกัน
- workflow GitHub Actions `day20.yml` (อยู่ที่ root ของ repo ตามที่ GitHub อ่าน)

## ผลที่รันจริง
- เทสต์ 66/66 ใน 12 ไฟล์ (Day 19 มี 64 ใน 11 ไฟล์)
- ทำให้พัง 3 แบบกับ app.config: จับได้ครบ แต่แบบ "ถอด provideHttpClient" จับได้ด้วยเทสต์นับ provider เท่านั้น เทสต์ "inject ClaimService ไม่ error" ผ่านเสมอเพราะ `provideHttpClientTesting()` ให้ HttpClient เอง จึงลบทิ้ง
- budget: ตั้ง error เหลือ 250kB build ล้ม exit 1 คืนค่าเดิม build ผ่าน
- `npm ci` ใหม่ + `npm run ci` ใน VM Linux ผ่าน (ใกล้เคียงขั้นตอนใน workflow แต่ไม่ใช่ runner ของ GitHub)

## เทียบกับที่รู้
- budget ใน angular.json คล้าย bundle size limit ที่ตั้งใน vite/webpack (size-limit) แต่ CLI เช็คให้ตอน build เลย
- เทสต์ที่ใช้ providers ตัวจริงของแอป (แทนการสร้าง config เองในเทสต์) คือวิธีจับของที่ "ลืมลงทะเบียน" ได้ตรงกว่า ถ้า mock เยอะเกินจะไม่เห็นข้อผิดพลาดนี้

## ที่ยังค้าง
- ผลบนหน้า Actions จริง (ยังไม่ push)
- ทดสอบบน Windows, backend จริง, Signal Forms, interceptor, SSR, `ng update`
