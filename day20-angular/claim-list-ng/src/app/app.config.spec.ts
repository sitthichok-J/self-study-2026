import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { RouterTestingHarness } from '@angular/router/testing';
import { appConfig } from './app.config';
import { ClaimDetail } from './claim-detail/claim-detail';
import { CLAIMS_URL } from './claim.service';
import { TEST_CLAIMS } from './claim.fixtures';

// Day 20: อุดช่องโหว่ที่จดไว้ใน Day 19 เทสต์ routes สร้าง config ของตัวเอง จึงไม่จับกรณี app.config.ts ลืม withComponentInputBinding()
// ไฟล์นี้ใช้ providers ของ appConfig จริง (เปลี่ยนเฉพาะ backend ของ HttpClient เป็นตัวทดสอบ)
function setup() {
  TestBed.resetTestingModule();
  TestBed.configureTestingModule({ providers: [...appConfig.providers, provideHttpClientTesting()] });
  return TestBed.inject(HttpTestingController);
}

describe('appConfig (providers ตัวจริง)', () => {
  it('เปิด withComponentInputBinding: :id ใน URL เข้า input id ของ ClaimDetail', async () => {
    const http = setup();
    const harness = await RouterTestingHarness.create();
    const detail = await harness.navigateByUrl('/claims/CLM-002', ClaimDetail);
    TestBed.tick();
    expect(detail.id()).toBe('CLM-002');
    http.expectOne(`${CLAIMS_URL}/CLM-002`).flush(TEST_CLAIMS[1]);
  });

  // provideHttpClientTesting ให้ HttpClient อยู่แล้ว เทสต์ "inject ClaimService ได้" จึงจับการลืม provideHttpClient() ไม่ได้ ใช้การนับ provider แทน
  it('baseline: จำนวน provider ใน appConfig (ถ้าเพิ่ม/ลดต้องตั้งใจ)', () => {
    expect(appConfig.providers.length).toBe(3);
  });
});
