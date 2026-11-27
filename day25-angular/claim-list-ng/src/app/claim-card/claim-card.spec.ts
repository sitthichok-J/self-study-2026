import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { ClaimCard } from './claim-card';
import { Claim } from '../claim';

const claim: Claim = {
  id: 'CLM-T1',
  patient_ref: 'FAKE-T1',
  amount: 99900, // สตางค์ = 999 บาท
  status: 'PENDING',
  created_at: '2026-11-18T01:00:00Z',
  updated_at: '2026-11-18T01:00:00Z',
};

async function setup(selected = false, override: Partial<Claim> = {}) {
  TestBed.resetTestingModule();
  await TestBed.configureTestingModule({ imports: [ClaimCard], providers: [provideRouter([])] }).compileComponents();
  const fixture = TestBed.createComponent(ClaimCard);
  fixture.componentRef.setInput('claim', { ...claim, ...override });
  fixture.componentRef.setInput('selected', selected);
  await fixture.whenStable();
  return fixture;
}

describe('ClaimCard', () => {
  it('แสดง id, patient_ref, ยอดเงินเป็นบาท (แปลงจากสตางค์) และสถานะ', async () => {
    const el = (await setup()).nativeElement as HTMLElement;
    expect(el.querySelector('.card-id')?.textContent).toBe('CLM-T1');
    expect(el.querySelector('.owner')?.textContent).toBe('FAKE-T1');
    expect(el.querySelector('.amount')?.textContent).toBe('999 บาท');
    expect(el.querySelector('.status')?.textContent).toBe('PENDING');
  });

  it('สตางค์ที่ไม่ลงตัวแสดงทศนิยม (76000 สตางค์ = 760 บาท, 12345 สตางค์ = 123.45 บาท)', async () => {
    const a = (await setup(false, { amount: 76000 })).nativeElement as HTMLElement;
    expect(a.querySelector('.amount')?.textContent).toBe('760 บาท');
    const b = (await setup(false, { amount: 12345 })).nativeElement as HTMLElement;
    expect(b.querySelector('.amount')?.textContent).toBe('123.45 บาท');
  });

  it('ส่ง event select พร้อม id เมื่อกดปุ่ม', async () => {
    const fixture = await setup();
    const emitted: string[] = [];
    fixture.componentInstance.select.subscribe((id) => emitted.push(id));
    (fixture.nativeElement as HTMLElement).querySelector('button')!.click();
    expect(emitted).toEqual(['CLM-T1']);
  });

  it('@switch แสดงป้ายสีตามสถานะ', async () => {
    const expected = { PENDING: 'pending', APPROVED: 'approved', REJECTED: 'rejected' } as const;
    for (const [status, cls] of Object.entries(expected)) {
      const fixture = await setup(false, { status: status as Claim['status'] });
      const badge = (fixture.nativeElement as HTMLElement).querySelector('.status')!;
      expect(badge.classList.contains(cls)).toBe(true);
      expect(badge.textContent).toBe(status);
    }
  });

  it('ใส่ class selected เมื่อ selected เป็น true เท่านั้น', async () => {
    const off = await setup(false);
    expect((off.nativeElement as HTMLElement).querySelector('.card')!.classList.contains('selected')).toBe(false);
    const on = await setup(true);
    expect((on.nativeElement as HTMLElement).querySelector('.card')!.classList.contains('selected')).toBe(true);
  });

  it('มีลิงก์ไปหน้ารายละเอียดของเคลมนี้ (/claims/:id)', async () => {
    const el = (await setup()).nativeElement as HTMLElement;
    expect(el.querySelector('a.detail-link')?.getAttribute('href')).toBe('/claims/CLM-T1');
  });
});
