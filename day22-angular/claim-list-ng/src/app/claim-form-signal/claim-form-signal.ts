import { Component, computed, inject, signal } from '@angular/core';
import { FieldTree, FormField, FormRoot, form, max, min, pattern, required, validate } from '@angular/forms/signals';
import { Router, RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { describeSubmitError, readApiError } from '../api-error';
import { MAX_AMOUNT_BAHT, PATIENT_REF_PATTERN, bahtToSatang } from '../claim';
import { ClaimService } from '../claim.service';

export interface ClaimFormModel {
  patient_ref: string;
  amountBaht: number | null;
}

// Signal Forms: ข้อมูลเก็บใน signal ธรรมดา (model) แล้ว form() ห่อเป็น field tree ที่อ่านสถานะเป็น signal
@Component({
  selector: 'app-claim-form-signal',
  imports: [FormField, FormRoot, RouterLink],
  styleUrl: '../claim-form/claim-form.css',
  templateUrl: './claim-form-signal.html',
})
export class ClaimFormSignal {
  private readonly service = inject(ClaimService);
  private readonly router = inject(Router);

  protected readonly model = signal<ClaimFormModel>({ patient_ref: '', amountBaht: null });
  protected readonly serverError = signal<string | null>(null);

  // Idempotency-Key ผูกกับค่าใน model: ข้อมูลเดิม = key เดิม (กดซ้ำ/retry ปลอดภัย) แก้ข้อมูลแล้วได้ key ใหม่
  // computed จำค่าไว้จนกว่า model จะเปลี่ยน จึงไม่ต้องมี effect/subscribe เหมือนฟอร์ม reactive
  private readonly idempotencyKey = computed(() => {
    this.model();
    return crypto.randomUUID();
  });

  protected readonly form = form(
    this.model,
    (p) => {
      required(p.patient_ref, { message: 'กรอก patient_ref' });
      pattern(p.patient_ref, PATIENT_REF_PATTERN, { message: 'ต้องเป็นรูปแบบ FAKE-XXXX (ตัวพิมพ์ใหญ่หรือตัวเลข 3-12 ตัว)' });
      required(p.amountBaht, { message: 'กรอกยอดเงิน' });
      min(p.amountBaht, 0.01, { message: 'ยอดเงินต้องมากกว่า 0' });
      validate(p.amountBaht, ({ value }) => {
        const v = value();
        return v === null || !hasMoreThanTwoDecimals(v) ? undefined : { kind: 'maxTwoDecimals', message: 'ทศนิยมได้ไม่เกิน 2 ตำแหน่ง' };
      });
      max(p.amountBaht, MAX_AMOUNT_BAHT, { message: `ยอดเงินต้องไม่เกิน ${MAX_AMOUNT_BAHT.toLocaleString('en-US')} บาท` });
    },
    {
      // [formRoot] เรียก action นี้เมื่อกดส่งและฟอร์มผ่าน validation ทุกช่อง
      submission: {
        action: async (f) => {
          this.serverError.set(null);
          const { patient_ref, amountBaht } = this.model();
          try {
            await firstValueFrom(
              this.service.create({ patient_ref, amount: bahtToSatang(amountBaht as number) }, this.idempotencyKey()),
            );
          } catch (err) {
            const message = describeSubmitError(err);
            const target = fieldFromServerMessage(f, readApiError(err));
            if (target) return { kind: 'server', message, fieldTree: target };
            this.serverError.set(message);
            return undefined;
          }
          void this.router.navigate(['/claims']);
          return undefined;
        },
      },
    },
  );

  // ข้อความ error ใต้ช่อง: แสดงเมื่อช่องถูกแตะแล้วหรือแก้แล้ว และเอาข้อความแรกตามลำดับที่ลงทะเบียนใน schema
  protected fieldError(field: FieldTree<unknown>): string | null {
    const state = field();
    if (!state.invalid() || !(state.touched() || state.dirty())) return null;
    return state.errors()[0]?.message ?? 'ค่าไม่ถูกต้อง';
  }
}

function hasMoreThanTwoDecimals(v: number): boolean {
  return Math.abs(v * 100 - Math.round(v * 100)) > 1e-7;
}

// backend ตอบ 400 ด้วยข้อความที่ระบุชื่อฟิลด์ (patient_ref / amount) จึงผูก error กลับเข้าช่องนั้น ไม่งั้นแสดงเป็น error รวม
function fieldFromServerMessage(
  f: FieldTree<ClaimFormModel>,
  api: { status: number; serverMessage: string | null } | null,
): FieldTree<unknown> | null {
  if (!api || api.status !== 400 || !api.serverMessage) return null;
  if (api.serverMessage.includes('patient_ref')) return f.patient_ref;
  if (api.serverMessage.includes('amount')) return f.amountBaht;
  return null;
}
