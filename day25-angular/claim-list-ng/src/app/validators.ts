import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

// validator ที่เขียนเอง: ยอดเงิน (บาท) ต้องไม่เกินวงเงิน
// คืน null เมื่อผ่าน คืน object เมื่อไม่ผ่าน (key คือชื่อ error ที่ template ใช้ดู)
export function maxBaht(limit: number): ValidatorFn {
  return (control: AbstractControl): ValidationErrors | null => {
    const value = control.value;
    if (value === null || value === '' || value === undefined) return null; // ให้ required จัดการช่องว่าง
    return Number(value) > limit ? { maxBaht: { limit, actual: Number(value) } } : null;
  };
}

// สตางค์เป็นจำนวนเต็ม ยอดที่มีทศนิยมเกิน 2 ตำแหน่งจะถูกปัดเป็นค่าอื่น จึงไม่ให้กรอก
export const maxTwoDecimals: ValidatorFn = (control: AbstractControl): ValidationErrors | null => {
  const value = control.value;
  if (value === null || value === '' || value === undefined) return null;
  return Math.abs(Number(value) * 100 - Math.round(Number(value) * 100)) > 1e-7 ? { maxTwoDecimals: true } : null;
};
