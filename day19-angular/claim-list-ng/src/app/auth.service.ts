import { Injectable, signal } from '@angular/core';

// จำลองสถานะล็อกอิน (ยังไม่มีระบบ auth จริง) เก็บเป็น signal ให้ guard กับ navbar อ่านค่าเดียวกัน
@Injectable({ providedIn: 'root' })
export class AuthService {
  readonly loggedIn = signal(false);

  login(): void {
    this.loggedIn.set(true);
  }

  logout(): void {
    this.loggedIn.set(false);
  }
}
