import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { NonNullableFormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MAX_AMOUNT_BAHT, PATIENT_REF_PATTERN, bahtToSatang } from '../claim';
import { ClaimService } from '../claim.service';
import { maxBaht, maxTwoDecimals } from '../validators';

export type FieldName = 'patient_ref' | 'amountBaht';

@Component({
  selector: 'app-claim-form',
  imports: [ReactiveFormsModule, RouterLink],
  styleUrl: './claim-form.css',
  templateUrl: './claim-form.html',
})
export class ClaimForm {
  private readonly fb = inject(NonNullableFormBuilder);
  private readonly service = inject(ClaimService);
  private readonly router = inject(Router);

  // reactive forms: โครงฟอร์มอยู่ใน class (ไม่ใช่ใน template เหมือน template-driven)
  protected readonly form = this.fb.group({
    patient_ref: ['', [Validators.required, Validators.pattern(PATIENT_REF_PATTERN)]],
    amountBaht: [null as number | null, [Validators.required, Validators.min(0.01), maxTwoDecimals, maxBaht(MAX_AMOUNT_BAHT)]],
  });

  protected readonly submitting = signal(false);
  protected readonly serverError = signal<string | null>(null);

  // Idempotency-Key สร้างครั้งเดียวต่อฟอร์ม: กดซ้ำหรือ retry หลัง error จะส่ง key เดิม backend จึงไม่สร้างซ้ำ
  // (เปลี่ยนค่าในฟอร์มแล้วส่งใหม่ต้องใช้ key ใหม่ ไม่งั้น backend ตอบ 409) จึงรีเซ็ต key เมื่อค่าเปลี่ยน
  private idempotencyKey = crypto.randomUUID();

  constructor() {
    this.form.valueChanges.subscribe(() => {
      this.idempotencyKey = crypto.randomUUID();
    });
  }

  // ข้อความ error ใต้ช่อง แสดงเมื่อช่องถูกแตะแล้วหรือกดส่งแล้ว
  fieldError(name: FieldName): string | null {
    const control = this.form.controls[name];
    if (!control.invalid || !(control.touched || control.dirty)) return null;
    const e = control.errors ?? {};
    if (e['required']) return name === 'patient_ref' ? 'กรอก patient_ref' : 'กรอกยอดเงิน';
    if (e['pattern']) return 'ต้องเป็นรูปแบบ FAKE-XXXX (ตัวพิมพ์ใหญ่หรือตัวเลข 3-12 ตัว)';
    if (e['min']) return 'ยอดเงินต้องมากกว่า 0';
    if (e['maxTwoDecimals']) return 'ทศนิยมได้ไม่เกิน 2 ตำแหน่ง';
    if (e['maxBaht']) return `ยอดเงินต้องไม่เกิน ${MAX_AMOUNT_BAHT.toLocaleString('en-US')} บาท`;
    return 'ค่าไม่ถูกต้อง';
  }

  submit(): void {
    if (this.submitting()) return; // กันกดซ้ำระหว่างรอ response
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const { patient_ref, amountBaht } = this.form.getRawValue();
    this.submitting.set(true);
    this.serverError.set(null);
    this.service.create({ patient_ref, amount: bahtToSatang(amountBaht as number) }, this.idempotencyKey).subscribe({
      next: () => {
        this.submitting.set(false);
        void this.router.navigate(['/claims']);
      },
      error: (err: unknown) => {
        this.submitting.set(false);
        this.serverError.set(describeSubmitError(err));
      },
    });
  }
}

function describeSubmitError(err: unknown): string {
  if (err instanceof HttpErrorResponse) {
    if (err.status === 0) return 'เชื่อมต่อ backend ไม่ได้';
    const msg = (err.error as { error?: string } | null)?.error;
    return msg ? `backend ปฏิเสธ (HTTP ${err.status}): ${msg}` : `backend ตอบ HTTP ${err.status}`;
  }
  return 'ส่งเคลมไม่สำเร็จ';
}
