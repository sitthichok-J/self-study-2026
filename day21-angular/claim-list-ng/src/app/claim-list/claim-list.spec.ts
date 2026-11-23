import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, TestRequest, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { vi } from 'vitest';
import { ClaimList } from './claim-list';
import { CLAIMS_URL } from '../claim.service';
import { TEST_RESPONSE } from '../claim.fixtures';

type Fixture = ComponentFixture<ClaimList>;

// สร้าง component โดยยังไม่ตอบ request เพื่อให้เทสต์สถานะ loading ได้
function create(): { fixture: Fixture; http: HttpTestingController; req: () => TestRequest } {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({ imports: [ClaimList], providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])] });
  const http = TestBed.inject(HttpTestingController);
  const fixture = TestBed.createComponent(ClaimList);
  return { fixture, http, req: () => http.expectOne((r) => r.url === CLAIMS_URL) };
}

async function setupLoaded(): Promise<Fixture> {
  const { fixture, req } = create();
  req().flush(TEST_RESPONSE);
  await fixture.whenStable();
  return fixture;
}

const el = (f: Fixture) => f.nativeElement as HTMLElement;
const ids = (f: Fixture) => Array.from(el(f).querySelectorAll('.card-id')).map((e) => e.textContent!.trim());

async function setStatus(f: Fixture, value: string) {
  const select = el(f).querySelector('.status-filter') as HTMLSelectElement;
  select.value = value;
  select.dispatchEvent(new Event('change'));
  await f.whenStable();
}

async function setSearch(f: Fixture, value: string) {
  const input = el(f).querySelector('.search') as HTMLInputElement;
  input.value = value;
  input.dispatchEvent(new Event('input'));
  await f.whenStable();
}

describe('ClaimList: โหลดข้อมูลจาก backend (วันที่ 3)', () => {
  it('ระหว่างรอ response แสดง loading และยังไม่มีการ์ด', async () => {
    const { fixture, req } = create();
    await fixture.whenStable();
    expect(el(fixture).querySelector('.loading')?.textContent).toContain('กำลังโหลด');
    expect(el(fixture).querySelectorAll('app-claim-card').length).toBe(0);
    req().flush(TEST_RESPONSE);
  });

  it('ได้ response แล้วซ่อน loading และแสดงการ์ดครบ 5 รายการ', async () => {
    const f = await setupLoaded();
    expect(el(f).querySelector('.loading')).toBeNull();
    expect(el(f).querySelector('.error')).toBeNull();
    expect(el(f).querySelectorAll('app-claim-card').length).toBe(5);
  });

  it('backend ตอบ 500 แสดงข้อความ error พร้อมปุ่มลองใหม่ และไม่แสดงการ์ด', async () => {
    const { fixture, req } = create();
    req().flush('boom', { status: 500, statusText: 'Server Error' });
    await fixture.whenStable();
    expect(el(fixture).querySelector('.error')?.textContent).toContain('backend ตอบ HTTP 500');
    expect(el(fixture).querySelector('.retry')).toBeTruthy();
    expect(el(fixture).querySelectorAll('app-claim-card').length).toBe(0);
  });

  it('เชื่อมต่อ backend ไม่ได้ (status 0) แสดงข้อความเฉพาะ', async () => {
    const { fixture, req } = create();
    req().error(new ProgressEvent('error'));
    await fixture.whenStable();
    expect(el(fixture).querySelector('.error')?.textContent).toContain('เชื่อมต่อ backend ไม่ได้');
  });

  it('กดลองใหม่แล้วยิง request ใหม่ และแสดงข้อมูลเมื่อสำเร็จ', async () => {
    const { fixture, http, req } = create();
    req().flush('x', { status: 503, statusText: 'Unavailable' });
    await fixture.whenStable();
    (el(fixture).querySelector('.retry') as HTMLButtonElement).click();
    await fixture.whenStable();
    expect(el(fixture).querySelector('.loading')).toBeTruthy(); // กลับไป loading ระหว่างรอ
    http.expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await fixture.whenStable();
    expect(el(fixture).querySelector('.error')).toBeNull();
    expect(el(fixture).querySelectorAll('app-claim-card').length).toBe(5);
  });

  it('ยอดเงินแปลงจากสตางค์เป็นบาท (รวม 5,156,000 สตางค์ = 51560 บาท)', async () => {
    const f = await setupLoaded();
    expect(el(f).querySelector('.total')!.textContent!.trim()).toBe('ยอดรวมทั้งหมด: 51560 บาท');
  });
});

describe('ClaimList: ตัวกรอง ค้นหา และเลือก (ของวันที่ 17 ยังทำงานบนข้อมูลจาก backend)', () => {
  it('เริ่มต้นยังไม่ได้เลือกรายการ แล้วกดเลือกการ์ดที่สองได้', async () => {
    const f = await setupLoaded();
    expect(el(f).querySelector('.selected-info')!.textContent?.trim()).toBe('ยังไม่ได้เลือกรายการ');
    (el(f).querySelectorAll('app-claim-card button')[1] as HTMLButtonElement).click();
    await f.whenStable();
    expect(el(f).querySelector('.selected-info')!.textContent?.trim()).toBe('เลือกอยู่: CLM-002');
    expect(el(f).querySelectorAll('.card.selected').length).toBe(1);
  });

  it('กรองตามสถานะ APPROVED เหลือ CLM-002 กับ CLM-005 แล้วกลับ ALL ได้ครบ', async () => {
    const f = await setupLoaded();
    await setStatus(f, 'APPROVED');
    expect(ids(f)).toEqual(['CLM-002', 'CLM-005']);
    await setStatus(f, 'ALL');
    expect(ids(f).length).toBe(5);
  });

  it('ค้นหาด้วย patient_ref และ id ไม่สนตัวพิมพ์เล็กใหญ่', async () => {
    const f = await setupLoaded();
    await setSearch(f, ' fake-gh ');
    expect(ids(f)).toEqual(['CLM-004']);
    await setSearch(f, 'clm-001');
    expect(ids(f)).toEqual(['CLM-001']);
  });

  it('ค้นหาไม่พบแล้วขึ้นข้อความ @empty และยอดรวมที่แสดงเป็น 0', async () => {
    const f = await setupLoaded();
    await setSearch(f, 'ไม่มีใครชื่อนี้');
    expect(el(f).querySelector('.empty')?.textContent?.trim()).toBe('ไม่พบรายการที่ตรงกับเงื่อนไข');
    expect(el(f).querySelector('.result-info')!.textContent!.trim()).toBe('แสดง 0 จาก 5 รายการ (ยอดรวมที่แสดง 0 บาท)');
  });

  it('ข้อความจำนวนและยอดรวมที่แสดงเปลี่ยนตามตัวกรอง (APPROVED = 44500 บาท)', async () => {
    const f = await setupLoaded();
    await setStatus(f, 'APPROVED');
    expect(el(f).querySelector('.result-info')!.textContent!.trim()).toBe('แสดง 2 จาก 5 รายการ (ยอดรวมที่แสดง 44500 บาท)');
  });

  it('effect log เมื่อตัวกรองเปลี่ยน', async () => {
    const log = vi.spyOn(console, 'log').mockImplementation(() => {});
    const f = await setupLoaded();
    log.mockClear();
    await setStatus(f, 'REJECTED');
    expect(log).toHaveBeenCalledWith('[claim-list] filter เปลี่ยน: status=REJECTED search=""');
    log.mockRestore();
  });

  // baseline วันที่ 3: เก็บพฤติกรรมของหน้ารายการที่ได้ข้อมูลจาก backend (ถ้าตั้งใจเปลี่ยน ให้รัน ng test -u)
  it('baseline: ข้อความที่หน้ารายการแสดง และผลกรองแต่ละสถานะ', async () => {
    const f = await setupLoaded();
    const cards = Array.from(el(f).querySelectorAll('app-claim-card')).map((c) => ({
      id: c.querySelector('.card-id')?.textContent,
      owner: c.querySelector('.owner')?.textContent,
      amount: c.querySelector('.amount')?.textContent,
      status: c.querySelector('.status')?.textContent,
    }));
    const byStatus: Record<string, string[]> = {};
    for (const status of ['ALL', 'PENDING', 'APPROVED', 'REJECTED']) {
      await setStatus(f, status);
      byStatus[status] = ids(f);
    }
    expect({ heading: el(f).querySelector('h2')?.textContent, cards, byStatus }).toMatchSnapshot();
  });
});
