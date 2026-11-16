import { TestBed } from '@angular/core/testing';
import { ClaimList } from './claim-list';
import { MOCK_CLAIMS } from '../claim';

async function setup() {
  await TestBed.configureTestingModule({ imports: [ClaimList] }).compileComponents();
  const fixture = TestBed.createComponent(ClaimList);
  await fixture.whenStable();
  return fixture;
}

describe('ClaimList', () => {
  it('แสดงการ์ดครบทุกรายการ (5 รายการ)', async () => {
    const fixture = await setup();
    const cards = (fixture.nativeElement as HTMLElement).querySelectorAll('app-claim-card');
    expect(cards.length).toBe(MOCK_CLAIMS.length);
    expect(cards.length).toBe(5);
  });

  it('เริ่มต้นยังไม่ได้เลือกรายการ', async () => {
    const fixture = await setup();
    const info = (fixture.nativeElement as HTMLElement).querySelector('.selected-info')!;
    expect(info.textContent?.trim()).toBe('ยังไม่ได้เลือกรายการ');
  });

  it('กดเลือกการ์ดแล้วแสดงรายการที่เลือกและใส่ class selected', async () => {
    const fixture = await setup();
    const root = fixture.nativeElement as HTMLElement;
    const buttons = root.querySelectorAll('app-claim-card button');
    (buttons[1] as HTMLButtonElement).click();
    await fixture.whenStable();
    expect(root.querySelector('.selected-info')!.textContent?.trim()).toBe('เลือกอยู่: CLM-002');
    expect(root.querySelectorAll('.card.selected').length).toBe(1);
  });

  // baseline: เก็บพฤติกรรมของหน้ารายการปัจจุบันไว้เป็น snapshot (ถ้าตั้งใจเปลี่ยน ให้รัน ng test -u)
  it('baseline: ข้อความที่หน้ารายการแสดง', async () => {
    const fixture = await setup();
    const root = fixture.nativeElement as HTMLElement;
    const summary = Array.from(root.querySelectorAll('app-claim-card')).map((c) => ({
      id: c.querySelector('.card-id')?.textContent,
      owner: c.querySelector('.owner')?.textContent,
      amount: c.querySelector('.amount')?.textContent,
      status: c.querySelector('.status')?.textContent,
    }));
    expect({ heading: root.querySelector('h2')?.textContent, summary }).toMatchSnapshot();
  });
});
