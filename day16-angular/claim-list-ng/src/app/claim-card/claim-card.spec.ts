import { TestBed } from '@angular/core/testing';
import { ClaimCard } from './claim-card';
import { Claim } from '../claim';

const claim: Claim = { id: 'CLM-T1', ownerName: 'ผู้ทดสอบ', amount: 999, status: 'PENDING' };

async function setup(selected = false) {
  await TestBed.configureTestingModule({ imports: [ClaimCard] }).compileComponents();
  const fixture = TestBed.createComponent(ClaimCard);
  fixture.componentRef.setInput('claim', claim);
  fixture.componentRef.setInput('selected', selected);
  await fixture.whenStable();
  return fixture;
}

describe('ClaimCard', () => {
  it('แสดง id ชื่อเจ้าของ ยอดเงิน และสถานะ', async () => {
    const fixture = await setup();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('.card-id')?.textContent).toBe('CLM-T1');
    expect(el.querySelector('.owner')?.textContent).toBe('ผู้ทดสอบ');
    expect(el.querySelector('.amount')?.textContent).toBe('999 บาท');
    expect(el.querySelector('.status')?.textContent).toBe('PENDING');
  });

  it('ส่ง event select พร้อม id เมื่อกดปุ่ม', async () => {
    const fixture = await setup();
    const emitted: string[] = [];
    fixture.componentInstance.select.subscribe((id) => emitted.push(id));
    (fixture.nativeElement as HTMLElement).querySelector('button')!.click();
    expect(emitted).toEqual(['CLM-T1']);
  });

  it('ใส่ class selected เมื่อ selected เป็น true เท่านั้น', async () => {
    const off = await setup(false);
    expect((off.nativeElement as HTMLElement).querySelector('.card')!.classList.contains('selected')).toBe(false);
    TestBed.resetTestingModule();
    const on = await setup(true);
    expect((on.nativeElement as HTMLElement).querySelector('.card')!.classList.contains('selected')).toBe(true);
  });
});
