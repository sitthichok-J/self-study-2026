# Day 22 — โน้ตวันที่ 7 (Signal Forms)

## ทำอะไร
- เขียนฟอร์มสร้างเคลมซ้ำด้วย Signal Forms ที่ `/claims/new-signal` เก็บฟอร์ม reactive เดิมไว้ที่ `/claims/new` เพื่อเทียบโค้ดและพฤติกรรมตรง ๆ
- เขียน parity spec ป้อนค่าเดียวกันเข้าสองฟอร์ม ต้องได้ข้อความ error และ POST body เท่ากัน
- ใช้ model เป็น `signal({patient_ref, amountBaht})` แล้วให้ `form(model, schema, {submission:{action}})` ผูก validator กับ field

## เทียบ reactive forms กับ Signal Forms (ฟอร์มนี้เท่านั้น ไม่ใช่ข้อสรุปทั่วไป)
| เรื่อง | reactive (`claim-form`) | Signal Forms (`claim-form-signal`) |
|---|---|---|
| บรรทัด ts / html | 74 / 30 | 93 / 30 |
| spec | 171 บรรทัด | 204 บรรทัด |
| validator | `Validators` + custom `maxBaht`, `maxTwoDecimals` ในไฟล์ validators | `required/pattern/min/max/validate` ใน schema เดียว พร้อมข้อความ error ในที่เดียวกัน |
| ค่าฟอร์ม | อยู่ใน FormGroup ต้องดึงด้วย `getRawValue()` | model เป็น signal ที่เราถืออยู่เอง |
| template | `formControlName` + `ReactiveFormsModule` | `[formField]` + `[formRoot]` |
| submit | handler เขียนเอง | `submission.action` และ state `submitting()` มาให้ |
| error จาก backend | set error ที่ control เอง | action คืน `{kind, message, fieldTree}` ไปแปะที่ field |

จำนวนบรรทัดของ Signal Forms มากกว่า เพราะฟอร์มนี้ผมใส่ logic map error backend เข้าช่อง และ key ที่ผูกกับ model ลงไปใน component ด้วย ส่วนของ reactive ที่เทียบกันไม่ได้ทำสองอย่างนี้ในไฟล์เดียว ตัวเลขนี้จึงไม่ได้บอกว่าแบบไหนสั้นกว่า

## ข้อดี ข้อเสีย (จากฟอร์มนี้ฟอร์มเดียว ไม่ได้ลองกับ FormArray, ฟอร์มซ้อนหลายชั้น หรือ custom control)

**Signal Forms**
- ดี: model เป็น signal ธรรมดา อ่านค่าและ state (`invalid()`, `touched()`, `errors()`) เป็น signal ใช้กับ OnPush/zoneless และ `computed` ได้ตรง ๆ ไม่ต้อง subscribe `valueChanges`
- ดี: validator และข้อความ error รวมอยู่ใน schema ที่เดียว อ่านจบในไฟล์เดียว
- ดี: `submission.action` ให้ `submitting()` และกัน submit ซ้อนมาให้ ใส่ error จาก backend ลงช่องผ่าน `fieldTree` ได้
- เสีย: API ใหม่ ตัวอย่าง/บทความ/คนที่เคยใช้ยังน้อยกว่า reactive มาก ต้องอ่าน d.ts เองหลายจุด
- เสีย: ผมยังไม่ได้ลองเรื่องที่ฟอร์มจริงมักเจอ (รายการที่เพิ่ม/ลบแถวได้, validator ข้าม field, async validator, ผูกกับ UI library) จึงไม่รู้ว่ารองรับดีแค่ไหน
- เสีย: เทสต์ async ต้องรอเพิ่ม (`settle()`) และ helper ของ reactive ใช้ต่อไม่ได้

**Reactive forms**
- ดี: ใช้กันแพร่หลาย เอกสารและ library รอบข้าง (UI kit, ตัวอย่าง) รองรับเต็ม มี FormArray, async validator ครบและเป็นที่รู้จัก
- ดี: ฟอร์มใหญ่ ๆ ที่มีอยู่แล้วไม่ต้องเขียนใหม่ ทีมส่วนใหญ่ก็อ่านออก
- เสีย: ค่าและ state เป็น Observable/getter ผสม ต้องแปลงเป็น signal เองถ้าอยากใช้กับ `computed`
- เสีย: validator กระจายอยู่ในไฟล์แยก ต้องกระโดดอ่าน และ type ของค่าไม่เข้มเท่า (ใช้ `NonNullableFormBuilder` ช่วยได้)

## แบบไหนเหมาะกับอะไร
- ฟอร์มใหม่ในแอปที่ใช้ signal/zoneless อยู่แล้ว และฟอร์มไม่ซับซ้อน (แบบฟอร์มสร้างเคลมนี้): Signal Forms น่าลอง เพราะเข้ากับส่วนอื่นของแอป
- ฟอร์มที่มีอยู่แล้ว หรือฟอร์มซับซ้อน (แถวไดนามิก, async validator, ผูก UI library): อยู่กับ reactive ก่อน
- คำตอบ "แบบไหนน่าใช้กว่า": ยังฟันธงไม่ได้ ฟอร์มนี้ผลออกมาพฤติกรรมเท่ากัน ต่างกันที่ความรู้สึกตอนเขียน ผมเอนไปทาง Signal Forms สำหรับฟอร์มใหม่ แต่ต้องลองฟอร์มซับซ้อนอีกอันก่อนเปลี่ยนจริง และรอดูว่าบนเครื่อง Windows/backend จริงยังเหมือนเดิมไหม

## ที่ได้เรียนรู้
- ความเสถียร: ดูจาก d.ts ที่ติดตั้ง API หลักเป็น `@publicApi 22.0` ส่วนหน้า angular.dev ที่อ่านไม่เห็นป้ายชัดเจน
- `submit` ที่ถูกเรียกซ้อนตอนกำลังส่งจะคืน false ไม่ยิง action ซ้ำ ปุ่มจึงผูก `submitting()` ได้เลย
- action เป็น async เทสต์ต้องรอให้จบ (ใช้ `settle()` = setTimeout 0 + `whenStable`) แค่ `whenStable` อย่างเดียวไม่พอ เจอเองตอนเทสต์ล้ม
- การ map error backend ไปช่องใช้ `includes('patient_ref'|'amount')` กับข้อความ เปราะ ควรให้ backend ส่ง field name เป็นโครงสร้างถ้าทำได้ (ตอนนี้ Day 8 backend ส่งแค่ `{"error": "..."}`)
- เทสต์ใช้ตัวช่วยของ reactive ไม่ได้ตรง ๆ ต้องเขียน helper ใหม่ จึงทำ parity spec เป็นตัวยึดว่าสองฟอร์มพฤติกรรมเท่ากัน

## เทียบกับที่รู้
- React Hook Form: `register` ผูก input กับ form state คล้าย `[formField]` และ schema ผ่าน resolver ส่วน Signal Forms ถือ schema ใน `form()` เอง ฝั่ง RHF จำจากที่เคยใช้ ยังไม่ได้ทดลองซ้ำวันนี้
- Vue: `ref` เป็น model แล้ว validate ด้วย library ภายนอก (เช่น vee-validate) Signal Forms ใกล้กับแบบนี้มากกว่า reactive forms ตรงที่ model เป็น signal

## ที่ยังค้าง
- ต่อ backend จริง ดูข้อความ 400 จริงว่า map ไปช่องได้ไหม (ดู README)
- bundle initial เพิ่ม ~4.36 kB จาก Day 21 ยังไม่รู้สาเหตุ ควรเปิด source-map-explorer ดู
- ตัดสินใจยังไม่ได้ว่าจะเลือกฟอร์มไหนเป็นตัวหลัก (ไว้ตัดสินหลังต่อ backend จริง และลองฟอร์มที่ซับซ้อนกว่านี้ เช่น แถวไดนามิก)
