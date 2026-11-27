import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, TestRequest, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, DeferBlockBehavior, DeferBlockState, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { vi } from 'vitest';
import { ClaimList } from './claim-list';
import { CLAIMS_URL } from '../claim.service';
import { TEST_CLAIMS, TEST_RESPONSE } from '../claim.fixtures';

type Fixture = ComponentFixture<ClaimList>;

// สร้าง component โดยยังไม่ตอบ request เพื่อให้เทสต์สถานะ loading ได้
function create(): { fixture: Fixture; http: HttpTestingController; req: () => TestRequest } {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({ imports: [ClaimList], providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])] });
  const http = TestBed.inject(HttpTestingController);
  const fixture = TestBed.createComponent(ClaimList);
  fixture.detectChanges(); // ให้ httpResource ยิง request ครั้งแรก
  return { fixture, http, req: () => http.expectOne((r) => r.url === CLAIMS_URL) };
}

// Day 24: pager อยู่ใน @defer (hydrate on interaction) → ในเทสต์ block นี้เรนเดอร์ตอน idle (หลัง whenStable) จึงต้อง detectChanges อีกรอบก่อนอ่าน DOM
async function settle(f: Fixture) {
  await f.whenStable();
  f.detectChanges();
  await f.whenStable();
  f.detectChanges();
}

async function setupLoaded(): Promise<Fixture> {
  const { fixture, req } = create();
  req().flush(TEST_RESPONSE);
  await settle(fixture);
  return fixture;
}

const el = (f: Fixture) => f.nativeElement as HTMLElement;
const ids = (f: Fixture) => Array.from(el(f).querySelectorAll('.card-id')).map((e) => e.textContent!.trim());

// วันที่ 23: เปลี่ยนสถานะแล้วหน้ารายการยิง request ใหม่พร้อม status ให้ backend กรอง (จำลอง backend ด้วยการ flush เฉพาะแถวที่ตรง)
async function setStatus(f: Fixture, value: string) {
  const http = TestBed.inject(HttpTestingController);
  const select = el(f).querySelector('.status-filter') as HTMLSelectElement;
  select.value = value;
  select.dispatchEvent(new Event('change'));
  f.detectChanges();
  TestBed.tick();
  const r = http.expectOne((x) => x.url === CLAIMS_URL);
  expect(r.request.params.get('status')).toBe(value === 'ALL' ? null : value);
  r.flush({ items: TEST_CLAIMS.filter((c) => value === 'ALL' || c.status === value), limit: 11, offset: 0 });
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
    fixture.detectChanges(); // ห้าม await whenStable: request ที่ค้างอยู่ของ httpResource นับเป็น pending task จะรอไม่จบ
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
    fixture.detectChanges();
    TestBed.tick();
    expect(el(fixture).querySelector('.loading')).toBeTruthy(); // กลับไป loading ระหว่างรอ
    http.expectOne((r) => r.url === CLAIMS_URL).flush(TEST_RESPONSE);
    await fixture.whenStable();
    expect(el(fixture).querySelector('.error')).toBeNull();
    expect(el(fixture).querySelectorAll('app-claim-card').length).toBe(5);
  });

  it('ยอดเงินแปลงจากสตางค์เป็นบาท (รวม 5,156,000 สตางค์ = 51560 บาท)', async () => {
    const f = await setupLoaded();
    expect(el(f).querySelector('.total')!.textContent!.trim()).toBe('ยอดรวมในหน้านี้: 51560 บาท');
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
    expect(el(f).querySelector('.result-info')!.textContent!.trim()).toBe('แสดง 0 จาก 5 รายการในหน้านี้ (ยอดรวมที่แสดง 0 บาท)');
  });

  it('ข้อความจำนวนและยอดรวมที่แสดงเปลี่ยนตามตัวกรอง (APPROVED = 44500 บาท)', async () => {
    const f = await setupLoaded();
    await setStatus(f, 'APPROVED');
    expect(el(f).querySelector('.result-info')!.textContent!.trim()).toBe('แสดง 2 จาก 2 รายการในหน้านี้ (ยอดรวมที่แสดง 44500 บาท)');
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
    byStatus['ALL'] = ids(f); // เริ่มที่ ALL อยู่แล้ว (ตั้งซ้ำค่าเดิมไม่ยิง request ใหม่)
    for (const status of ['PENDING', 'APPROVED', 'REJECTED']) {
      await setStatus(f, status);
      byStatus[status] = ids(f);
    }
    expect({ heading: el(f).querySelector('h2')?.textContent, cards, byStatus }).toMatchSnapshot();
  });
});

// วันที่ 23: แบ่งหน้า และ select ไม่เด้งกลับค่าเดิมหลังโหลดใหม่
describe('ClaimList: แบ่งหน้าและตัวกรองที่ส่งไป backend (วันที่ 23)', () => {
  const eleven = Array.from({ length: 11 }, (_, i) => ({ ...TEST_CLAIMS[0], id: `CLM-${String(i + 1).padStart(3, '0')}` }));
  const btn = (f: Fixture, c: string) => el(f).querySelector(c) as HTMLButtonElement;

  async function loadedWith(items: typeof eleven): Promise<Fixture> {
    const { fixture, req } = create();
    req().flush({ items, limit: 11, offset: 0 });
    await settle(fixture);
    return fixture;
  }

  it('มี 11 แถว: แสดงการ์ด 10 ใบ ปุ่มถัดไปกดได้ ปุ่มก่อนหน้ากดไม่ได้ และเขียนว่าหน้า 1', async () => {
    const f = await loadedWith(eleven);
    expect(el(f).querySelectorAll('app-claim-card').length).toBe(10);
    expect(btn(f, '.next').disabled).toBe(false);
    expect(btn(f, '.prev').disabled).toBe(true);
    expect(el(f).querySelector('.page-info')!.textContent!.trim()).toBe('หน้า 1');
  });

  it('select ขนาดหน้าเริ่มต้นแสดง 10 ตรงกับที่ store ใช้ขอ (ไม่ใช่ตัวเลือกแรก)', async () => {
    const f = await setupLoaded();
    expect((el(f).querySelector('.page-size') as HTMLSelectElement).value).toBe('10');
    // Day 24: ต้องเป็น attribute selected ใน HTML (server เขียน property ไม่ได้) พบจากภาพหน้าจอตอนปิด JavaScript
    expect(el(f).querySelector('.page-size option[selected]')?.getAttribute('value')).toBe('10');
  });

  it('มี 5 แถว (หน้าเดียว): ปุ่มถัดไปกดไม่ได้', async () => {
    const f = await setupLoaded();
    expect(btn(f, '.next').disabled).toBe(true);
  });

  it('กดถัดไป ยิง request offset 10 แล้วขึ้นหน้า 2 และปุ่มก่อนหน้ากดได้', async () => {
    const f = await loadedWith(eleven);
    const http = TestBed.inject(HttpTestingController);
    btn(f, '.next').click();
    f.detectChanges();
    TestBed.tick();
    const r = http.expectOne((x) => x.url === CLAIMS_URL);
    expect(r.request.params.get('offset')).toBe('10');
    r.flush({ items: [{ ...TEST_CLAIMS[0], id: 'CLM-011' }], limit: 11, offset: 10 });
    await f.whenStable();
    expect(el(f).querySelector('.page-info')!.textContent!.trim()).toBe('หน้า 2');
    expect(ids(f)).toEqual(['CLM-011']);
    expect(btn(f, '.prev').disabled).toBe(false);
    expect(btn(f, '.next').disabled).toBe(true);
  });

  it('เลือกขนาดหน้า 5 แล้วขอ limit 6 และแสดงการ์ด 5 ใบ', async () => {
    const f = await loadedWith(eleven);
    const http = TestBed.inject(HttpTestingController);
    const sel = el(f).querySelector('.page-size') as HTMLSelectElement;
    sel.value = '5';
    sel.dispatchEvent(new Event('change'));
    f.detectChanges();
    TestBed.tick();
    const r = http.expectOne((x) => x.url === CLAIMS_URL);
    expect(r.request.params.get('limit')).toBe('6');
    r.flush({ items: eleven.slice(0, 6), limit: 6, offset: 0 });
    await f.whenStable();
    expect(el(f).querySelectorAll('app-claim-card').length).toBe(5);
    expect((el(f).querySelector('.page-size') as HTMLSelectElement).value).toBe('5');
  });

  it('เลือกสถานะแล้วโหลดเสร็จ select ยังค้างที่ค่าที่เลือก (ไม่เด้งกลับเป็นทั้งหมด)', async () => {
    const f = await setupLoaded();
    const before = el(f).querySelector('.status-filter');
    await setStatus(f, 'REJECTED');
    // ตัวกรองอยู่นอกกิ่ง @if (loading()) จึงไม่ถูกสร้างใหม่ระหว่างโหลด (ถ้าถูกสร้างใหม่ ค่าจะเด้งกลับเป็น ALL)
    expect(el(f).querySelector('.status-filter')).toBe(before);
    expect((el(f).querySelector('.status-filter') as HTMLSelectElement).value).toBe('REJECTED');
    expect(ids(f)).toEqual(['CLM-003']);
  });

  it('ค้นหาทำงานเฉพาะในหน้าที่โหลดมา และไม่ยิง request ใหม่', async () => {
    const f = await loadedWith(eleven);
    const http = TestBed.inject(HttpTestingController);
    await setSearch(f, 'clm-003');
    expect(ids(f)).toEqual(['CLM-003']);
    http.expectNone((x) => x.url === CLAIMS_URL);
  });
});

// Day 24: incremental hydration — pager ต้องอยู่ใน @defer block เดียว (ถ้าใครลบ @defer ออก ไม่มีเทสต์ DOM อื่นจับได้ เพราะ pager ยังทำงานเหมือนเดิม)
describe('ClaimList: pager ใน @defer (hydrate on interaction)', () => {
  function manual() {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [ClaimList],
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
      deferBlockBehavior: DeferBlockBehavior.Manual,
    });
    const fixture = TestBed.createComponent(ClaimList);
    fixture.detectChanges();
    return fixture;
  }

  it('มี defer block 1 ตัว และยังไม่มี pager จนกว่าจะ render block', async () => {
    const f = manual();
    const blocks = await f.getDeferBlocks();
    expect(blocks.length).toBe(1);
    expect(el(f).querySelector('.pager')).toBeNull();
    await blocks[0].render(DeferBlockState.Complete);
    expect(el(f).querySelector('.pager')).not.toBeNull();
    expect(el(f).querySelector('.pager .next')).not.toBeNull();
  });
});
