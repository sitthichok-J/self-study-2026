import { Component, signal } from '@angular/core';
import { ClaimCard } from '../claim-card/claim-card';
import { Claim, MOCK_CLAIMS } from '../claim';

@Component({
  selector: 'app-claim-list',
  // standalone component: component อื่นที่ใช้ใน template ต้องประกาศในอาเรย์ imports ของ component นี้เอง
  imports: [ClaimCard],
  styleUrl: './claim-list.css',
  templateUrl: './claim-list.html',
})
export class ClaimList {
  protected readonly claims = signal<Claim[]>(MOCK_CLAIMS);
  protected readonly selectedId = signal<string | null>(null);

  onSelect(id: string): void {
    this.selectedId.set(id);
  }
}
