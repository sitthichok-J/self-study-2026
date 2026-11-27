import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { vi } from 'vitest';
import { ClaimFormSignal as ClaimForm } from './claim-form-signal';
import { CLAIMS_URL } from '../claim.service';
import { TEST_CLAIMS } from '../claim.fixtures';

type Fixture = ComponentFixture<ClaimForm>;

async function setup() {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({ imports: [ClaimForm], providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])] });
  const fixture = TestBed.createComponent(ClaimForm);
  await fixture.whenStable();
  const navigate = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);
  return { fixture, navigate, http: TestBed.inject(HttpTestingController) };
}

const el = (f: Fixture) => f.nativeElement as HTMLElement;
const input = (f: Fixture, id: string) => el(f).querySelector(`#${id}`) as HTMLInputElement;
const errorOf = (f: Fixture, field: string) => el(f).querySelector(`[data-field="${field}"]`)?.textContent?.trim() ?? null;

async function type(f: Fixture, id: string, value: string, blur = true) {
  const i = input(f, id);
  i.value = value;
  i.dispatchEvent(new Event('input'));
  if (blur) i.dispatchEvent(new Event('blur'));
  await f.whenStable();
}

// action ของ Signal Forms เป็น async: หลัง flush response ต้องรอให้ promise ใน submit() จบก่อน (whenStable อย่างเดียวไม่พอ)
async function settle(f: Fixture) {
  await new Promise((r) => setTimeout(r, 0));
  await f.whenStable();
}

async function submit(f: Fixture) {
  (el(f).querySelector('form') as HTMLFormElement).dispatchEvent(new Event('submit'));
  await f.whenStable();
}

async function fillValid(f: Fixture, ref = 'FAKE-AB12', baht = '1500') {
  await type(f, 'patient_ref', ref);
  await type(f, 'amountBaht', baht);
}

describe('ClaimFormSignal (Signal Forms)', () => {
  it('เริ่มต้นไม่แสดง error สักช่อง แม้ฟอร์มยังไม่ผ่าน', async () => {
    const { fixture } = await setup();
    expect(errorOf(fixture, 'patient_ref')).toBeNull();
    expect(errorOf(fixture, 'amountBaht')).toBeNull();
  });

  it('กดส่งตอนว่าง: ไม่ยิง POST และแสดง error ทั้งสองช่อง', async () => {
    const { fixture, http, navigate } = await setup();
    await submit(fixture);
    http.expectNone(CLAIMS_URL);
    expect(errorOf(fixture, 'patient_ref')).toBe('กรอก patient_ref');
    expect(errorOf(fixture, 'amountBaht')).toBe('กรอกยอดเงิน');
    expect(navigate).not.toHaveBeenCalled();
  });

  it('patient_ref ผิดรูปแบบ (ไม่ใช่ FAKE-XXXX) แสดง error หลังออกจากช่อง', async () => {
    const { fixture } = await setup();
    await type(fixture, 'patient_ref', '1234567890123');
    expect(errorOf(fixture, 'patient_ref')).toContain('FAKE-XXXX');
    await type(fixture, 'patient_ref', 'fake-ab12'); // ตัวพิมพ์เล็กไม่ผ่าน
    expect(errorOf(fixture, 'patient_ref')).toContain('FAKE-XXXX');
    await type(fixture, 'patient_ref', 'FAKE-AB'); // สั้นไป (ต้อง 3-12 ตัว)
    expect(errorOf(fixture, 'patient_ref')).toContain('FAKE-XXXX');
    await type(fixture, 'patient_ref', 'FAKE-AB1');
    expect(errorOf(fixture, 'patient_ref')).toBeNull();
  });

  it('ยอดเงิน: 0 และติดลบ → "ต้องมากกว่า 0", เกินวงเงิน → ข้อความวงเงิน, 3 ทศนิยม → ทศนิยมไม่เกิน 2', async () => {
    const { fixture } = await setup();
    await type(fixture, 'amountBaht', '0');
    expect(errorOf(fixture, 'amountBaht')).toBe('ยอดเงินต้องมากกว่า 0');
    await type(fixture, 'amountBaht', '-5');
    expect(errorOf(fixture, 'amountBaht')).toBe('ยอดเงินต้องมากกว่า 0');
    await type(fixture, 'amountBaht', '10000000.01');
    expect(errorOf(fixture, 'amountBaht')).toBe('ยอดเงินต้องไม่เกิน 10,000,000 บาท');
    await type(fixture, 'amountBaht', '1.005');
    expect(errorOf(fixture, 'amountBaht')).toBe('ทศนิยมได้ไม่เกิน 2 ตำแหน่ง');
    await type(fixture, 'amountBaht', '10000000');
    expect(errorOf(fixture, 'amountBaht')).toBeNull();
  });

  it('ค่าถูก: POST body เป็นสตางค์ + header Idempotency-Key แล้วกลับ /claims', async () => {
    const { fixture, http, navigate } = await setup();
    await fillValid(fixture, 'FAKE-AB12', '19.99');
    await submit(fixture);
    const req = http.expectOne(CLAIMS_URL);
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ patient_ref: 'FAKE-AB12', amount: 1999 });
    expect(req.request.headers.get('Idempotency-Key')).toMatch(/^[0-9a-f-]{36}$/);
    req.flush(TEST_CLAIMS[0], { status: 201, statusText: 'Created' });
    await settle(fixture);
    expect(navigate).toHaveBeenCalledWith(['/claims']);
  });

  it('ระหว่างรอ response ปุ่มถูกปิด กดซ้ำไม่ยิงซ้ำ', async () => {
    const { fixture, http } = await setup();
    await fillValid(fixture);
    await submit(fixture);
    const btn = el(fixture).querySelector('button.submit') as HTMLButtonElement;
    expect(btn.disabled).toBe(true);
    expect(btn.textContent).toContain('กำลังส่ง');
    await submit(fixture);
    http.expectOne(CLAIMS_URL).flush(TEST_CLAIMS[0], { status: 201, statusText: 'Created' }); // expectOne = ต้องมีแค่ 1
  });

  it('backend ปฏิเสธ (400) ที่ระบุชื่อฟิลด์: ข้อความไปอยู่ใต้ช่องนั้น ไม่ออกจากหน้า และปุ่มกลับมากดได้', async () => {
    const { fixture, http, navigate } = await setup();
    await fillValid(fixture);
    await submit(fixture);
    http.expectOne(CLAIMS_URL).flush({ error: 'invalid input: amount must be between 1 and 1000000000' }, { status: 400, statusText: 'Bad Request' });
    await settle(fixture);
    expect(errorOf(fixture, 'amountBaht')).toContain('HTTP 400');
    expect(errorOf(fixture, 'amountBaht')).toContain('amount must be between');
    expect(errorOf(fixture, 'patient_ref')).toBeNull();
    expect(el(fixture).querySelector('.server-error')).toBeNull();
    expect((el(fixture).querySelector('button.submit') as HTMLButtonElement).disabled).toBe(false);
    expect(navigate).not.toHaveBeenCalled();
  });

  it('backend ปฏิเสธ (400) ที่ระบุ patient_ref: ข้อความไปอยู่ใต้ช่อง patient_ref', async () => {
    const { fixture, http } = await setup();
    await fillValid(fixture);
    await submit(fixture);
    http.expectOne(CLAIMS_URL).flush({ error: 'invalid input: patient_ref must match FAKE-[A-Z0-9]{3,12}' }, { status: 400, statusText: 'Bad Request' });
    await settle(fixture);
    expect(errorOf(fixture, 'patient_ref')).toContain('patient_ref must match');
    expect(errorOf(fixture, 'amountBaht')).toBeNull();
  });

  it('backend ปฏิเสธ (409 key ซ้ำคนละ body) ไม่ระบุฟิลด์: แสดงเป็นข้อความรวม .server-error', async () => {
    const { fixture, http } = await setup();
    await fillValid(fixture);
    await submit(fixture);
    http.expectOne(CLAIMS_URL).flush({ error: 'Idempotency-Key was already used with a different request' }, { status: 409, statusText: 'Conflict' });
    await settle(fixture);
    expect(el(fixture).querySelector('.server-error')?.textContent).toContain('HTTP 409');
    expect(errorOf(fixture, 'patient_ref')).toBeNull();
    expect(errorOf(fixture, 'amountBaht')).toBeNull();
  });

  it('ต่อไม่ติด (status 0) แสดง "เชื่อมต่อ backend ไม่ได้"', async () => {
    const { fixture, http } = await setup();
    await fillValid(fixture);
    await submit(fixture);
    http.expectOne(CLAIMS_URL).error(new ProgressEvent('error'));
    await settle(fixture);
    expect(el(fixture).querySelector('.server-error')?.textContent).toContain('เชื่อมต่อ backend ไม่ได้');
  });

  it('retry ด้วยข้อมูลเดิมใช้ Idempotency-Key เดิม แต่แก้ข้อมูลแล้วใช้ key ใหม่', async () => {
    const { fixture, http } = await setup();
    await fillValid(fixture);
    await submit(fixture);
    const first = http.expectOne(CLAIMS_URL);
    const key1 = first.request.headers.get('Idempotency-Key');
    first.flush('x', { status: 503, statusText: 'Unavailable' });
    await settle(fixture);

    await submit(fixture); // ข้อมูลเดิม
    const second = http.expectOne(CLAIMS_URL);
    expect(second.request.headers.get('Idempotency-Key')).toBe(key1);
    second.flush('x', { status: 503, statusText: 'Unavailable' });
    await settle(fixture);

    await type(fixture, 'amountBaht', '2000'); // แก้ข้อมูล
    await submit(fixture);
    const third = http.expectOne(CLAIMS_URL);
    expect(third.request.headers.get('Idempotency-Key')).not.toBe(key1);
    expect(third.request.body.amount).toBe(200000);
    third.flush(TEST_CLAIMS[0], { status: 201, statusText: 'Created' });
  });

  it('baseline: ข้อความ error ของแต่ละค่าที่กรอก (เก็บพฤติกรรมปัจจุบันไว้)', async () => {
    const { fixture } = await setup();
    // field tree กับ method ที่ protected: เข้าถึงผ่าน any เฉพาะในเทสต์นี้
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const c = fixture.componentInstance as any;
    const cases: [string, number | null][] = [
      ['', null],
      ['FAKE-AB12', 1500],
      ['fake-ab12', 0],
      ['FAKE-AB', -1],
      ['FAKE-AB12', 0.005],
      ['FAKE-AB12', 10_000_000.01],
    ];
    const table = [];
    for (const [ref, baht] of cases) {
      c.model.set({ patient_ref: ref, amountBaht: baht });
      c.form().markAsTouched();
      await fixture.whenStable();
      table.push({ ref, baht, patient_ref: c.fieldError(c.form.patient_ref), amountBaht: c.fieldError(c.form.amountBaht) });
    }
    expect(table).toMatchSnapshot();
  });
});
