import { Component, input, output } from '@angular/core';
import { Claim } from '../claim';

@Component({
  selector: 'app-claim-card',
  styleUrl: './claim-card.css',
  templateUrl: './claim-card.html',
})
export class ClaimCard {
  // รับข้อมูลเข้าด้วย input() (เทียบ props ใน Vue/React) ต้องมีค่าเสมอ จึงใช้ input.required
  readonly claim = input.required<Claim>();
  readonly selected = input(false);

  // ส่ง event ออกด้วย output() (เทียบ emits ใน Vue)
  readonly select = output<string>();

  onSelect(): void {
    this.select.emit(this.claim().id);
  }
}
