import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
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

// ---------- วันที่ 2: กรอง ค้นหา ยอดรวม ----------
function ids(root: HTMLElement): string[] {
  return Array.from(root.querySelectorAll('.card-id')).map((e) => e.textContent!.trim());
}

async function setStatus(fixture: Awaited<ReturnType<typeof setup>>, value: string) {
  const select = (fixture.nativeElement as HTMLElement).querySelector('.status-filter') as HTMLSelectElement;
  select.value = value;
  select.dispatchEvent(new Event('change'));
  await fixture.whenStable();
}

async function setSearch(fixture: Awaited<ReturnType<typeof setup>>, value: string) {
  const input = (fixture.nativeElement as HTMLElement).querySelector('.search') as HTMLInputElement;
  input.value = value;
  input.dispatchEvent(new Event('input'));
  await fixture.whenStable();
}

describe('ClaimList วันที่ 2', () => {
  it('กรองตามสถานะ APPROVED เหลือ CLM-002 กับ CLM-005', async () => {
    const fixture = await setup();
    await setStatus(fixture, 'APPROVED');
    expect(ids(fixture.nativeElement)).toEqual(['CLM-002', 'CLM-005']);
  });

  it('กลับไปเลือก ALL แล้วแสดงครบ 5 รายการ', async () => {
    const fixture = await setup();
    await setStatus(fixture, 'REJECTED');
    expect(ids(fixture.nativeElement)).toEqual(['CLM-003']);
    await setStatus(fixture, 'ALL');
    expect(ids(fixture.nativeElement).length).toBe(5);
  });

  it('ค้นหาด้วยชื่อ (ไม่สนตัวพิมพ์เล็กใหญ่ของ id และตัดช่องว่างหัวท้าย)', async () => {
    const fixture = await setup();
    await setSearch(fixture, ' มานี ');
    expect(ids(fixture.nativeElement)).toEqual(['CLM-004']);
    await setSearch(fixture, 'clm-001');
    expect(ids(fixture.nativeElement)).toEqual(['CLM-001']);
  });

  it('กรองสถานะและค้นหารวมกันแบบ AND', async () => {
    const fixture = await setup();
    await setStatus(fixture, 'PENDING');
    await setSearch(fixture, 'ตัวอย่าง');
    expect(ids(fixture.nativeElement)).toEqual(['CLM-001', 'CLM-004']);
    await setSearch(fixture, 'ทดสอบ');
    expect(ids(fixture.nativeElement)).toEqual([]);
  });

  it('ไม่พบผลลัพธ์แล้วแสดงข้อความ @empty', async () => {
    const fixture = await setup();
    await setSearch(fixture, 'ไม่มีใครชื่อนี้');
    const root = fixture.nativeElement as HTMLElement;
    expect(root.querySelectorAll('app-claim-card').length).toBe(0);
    expect(root.querySelector('.empty')?.textContent?.trim()).toBe('ไม่พบรายการที่ตรงกับเงื่อนไข');
  });

  it('ยอดรวมและจำนวนต่อสถานะคิดจากทุกรายการ ไม่เปลี่ยนตามตัวกรอง', async () => {
    const fixture = await setup();
    const root = fixture.nativeElement as HTMLElement;
    const before = root.querySelector('.summary')!.textContent!.replace(/\s+/g, ' ').trim();
    expect(before).toContain('ยอดรวมทั้งหมด: 51560 บาท'); // 1500+32000+4800+760+12500
    expect(before).toContain('PENDING 2 · APPROVED 2 · REJECTED 1');
    await setStatus(fixture, 'APPROVED');
    expect(root.querySelector('.summary')!.textContent!.replace(/\s+/g, ' ').trim()).toBe(before);
  });

  it('ข้อความจำนวนและยอดรวมที่แสดงเปลี่ยนตามตัวกรอง', async () => {
    const fixture = await setup();
    const root = fixture.nativeElement as HTMLElement;
    expect(root.querySelector('.result-info')!.textContent!.trim()).toBe(
      'แสดง 5 จาก 5 รายการ (ยอดรวมที่แสดง 51560 บาท)',
    );
    await setStatus(fixture, 'APPROVED');
    expect(root.querySelector('.result-info')!.textContent!.trim()).toBe(
      'แสดง 2 จาก 5 รายการ (ยอดรวมที่แสดง 44500 บาท)',
    );
  });

  it('effect log เมื่อตัวกรองเปลี่ยน', async () => {
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    const fixture = await setup();
    log.mockClear();
    await setStatus(fixture, 'REJECTED');
    expect(log).toHaveBeenCalledWith('[claim-list] filter เปลี่ยน: status=REJECTED search=""');
    log.mockRestore();
  });

  // baseline วันที่ 2: พฤติกรรมของตัวกรองเก็บเป็น snapshot (ถ้าตั้งใจเปลี่ยน ให้รัน ng test -u)
  it('baseline: ผลกรองแต่ละสถานะ', async () => {
    const fixture = await setup();
    const result: Record<string, string[]> = {};
    for (const status of ['ALL', 'PENDING', 'APPROVED', 'REJECTED']) {
      await setStatus(fixture, status);
      result[status] = ids(fixture.nativeElement);
    }
    expect(result).toMatchSnapshot();
  });
});
