import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { vi } from 'vitest';
import { ClaimFormSignal } from './claim-form-signal/claim-form-signal';
import { ClaimForm } from './claim-form/claim-form';
import { CLAIMS_URL } from './claim.service';
import { TEST_CLAIMS } from './claim.fixtures';

// ฟอร์ม reactive (Day 19) กับ Signal Forms (Day 22) ต้องให้ผลเหมือนกันจากมุมผู้ใช้: ข้อความ error และ request ที่ส่ง
async function mount<T>(cmp: new () => T): Promise<ComponentFixture<T>> {
  const f = TestBed.createComponent(cmp);
  await f.whenStable();
  return f;
}
const el = (f: ComponentFixture<unknown>) => f.nativeElement as HTMLElement;

async function fill(f: ComponentFixture<unknown>, id: string, value: string) {
  const i = el(f).querySelector(`#${id}`) as HTMLInputElement;
  i.value = value;
  i.dispatchEvent(new Event('input'));
  i.dispatchEvent(new Event('blur'));
  await f.whenStable();
}

const errors = (f: ComponentFixture<unknown>) =>
  Array.from(el(f).querySelectorAll('[data-field]')).map((e) => `${e.getAttribute('data-field')}: ${e.textContent?.trim()}`);

describe('ฟอร์ม reactive กับ Signal Forms ให้ผลตรงกัน', () => {
  beforeEach(() => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({ providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])] });
    vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);
  });

  const inputs: [string, string][] = [
    ['', ''],
    ['FAKE-AB12', '1500'],
    ['fake-ab12', '0'],
    ['FAKE-AB', '-5'],
    ['FAKE-AB12', '1.005'],
    ['FAKE-AB12', '10000000.01'],
    ['FAKE-ABCDEFGHIJKLM', '19.99'],
  ];

  it.each(inputs)('ข้อความ error เหมือนกันเมื่อกรอก patient_ref=%j ยอด=%j', async (ref, baht) => {
    const a = await mount(ClaimForm);
    const b = await mount(ClaimFormSignal);
    for (const f of [a, b]) {
      await fill(f, 'patient_ref', ref);
      await fill(f, 'amountBaht', baht);
    }
    expect(errors(b)).toEqual(errors(a));
  });

  it('ค่าถูกต้อง: POST body เหมือนกัน', async () => {
    const http = TestBed.inject(HttpTestingController);
    const bodies: unknown[] = [];
    for (const cmp of [ClaimForm, ClaimFormSignal] as const) {
      const f = await mount(cmp as new () => unknown);
      await fill(f, 'patient_ref', 'FAKE-AB12');
      await fill(f, 'amountBaht', '19.99');
      (el(f).querySelector('form') as HTMLFormElement).dispatchEvent(new Event('submit'));
      await f.whenStable();
      const req = http.expectOne(CLAIMS_URL);
      bodies.push(req.request.body);
      req.flush(TEST_CLAIMS[0], { status: 201, statusText: 'Created' });
      await new Promise((r) => setTimeout(r, 0));
    }
    expect(bodies[1]).toEqual(bodies[0]);
    expect(bodies[0]).toEqual({ patient_ref: 'FAKE-AB12', amount: 1999 });
  });
});
