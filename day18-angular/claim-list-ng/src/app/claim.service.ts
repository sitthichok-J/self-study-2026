import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ClaimListResponse, ClaimStatus } from './claim';

export interface ClaimQuery {
  status?: ClaimStatus;
  limit?: number;
  offset?: number;
}

export const CLAIMS_URL = '/api/claims'; // dev proxy (proxy.conf.json) ตัด /api ออกแล้วส่งต่อไป backend :8080

@Injectable({ providedIn: 'root' })
export class ClaimService {
  // service ได้ HttpClient ผ่าน inject() ไม่ต้องสร้างเอง
  private readonly http = inject(HttpClient);

  list(query: ClaimQuery = {}): Observable<ClaimListResponse> {
    let params = new HttpParams().set('limit', query.limit ?? 100).set('offset', query.offset ?? 0);
    if (query.status) params = params.set('status', query.status);
    return this.http.get<ClaimListResponse>(CLAIMS_URL, { params });
  }
}
