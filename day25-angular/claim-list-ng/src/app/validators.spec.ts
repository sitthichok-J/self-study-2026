import { FormControl } from '@angular/forms';
import { maxBaht, maxTwoDecimals } from './validators';

describe('maxBaht (validator ที่เขียนเอง)', () => {
  const v = maxBaht(1000);
  it('ค่าเท่ากับหรือต่ำกว่าวงเงินผ่าน', () => {
    expect(v(new FormControl(1000))).toBeNull();
    expect(v(new FormControl(0.01))).toBeNull();
  });
  it('เกินวงเงินแม้ 1 สตางค์ไม่ผ่าน และบอก limit/actual', () => {
    expect(v(new FormControl(1000.01))).toEqual({ maxBaht: { limit: 1000, actual: 1000.01 } });
  });
  it('ช่องว่าง/null ไม่ใช่หน้าที่ของ validator นี้ (ให้ required ตัดสิน)', () => {
    expect(v(new FormControl(null))).toBeNull();
    expect(v(new FormControl(''))).toBeNull();
  });
});

describe('maxTwoDecimals', () => {
  it('จำนวนเต็มและทศนิยม 1-2 ตำแหน่งผ่าน (รวมค่าที่ floating point เพี้ยน เช่น 19.99, 0.07)', () => {
    for (const x of [1, 12.5, 12.34, 19.99, 0.07, 1.1]) expect(maxTwoDecimals(new FormControl(x))).toBeNull();
  });
  it('ทศนิยมตำแหน่งที่ 3 ไม่ผ่าน', () => {
    expect(maxTwoDecimals(new FormControl(1.005))).toEqual({ maxTwoDecimals: true });
    expect(maxTwoDecimals(new FormControl(0.001))).toEqual({ maxTwoDecimals: true });
  });
  it('ช่องว่างผ่าน', () => {
    expect(maxTwoDecimals(new FormControl(null))).toBeNull();
  });
});
