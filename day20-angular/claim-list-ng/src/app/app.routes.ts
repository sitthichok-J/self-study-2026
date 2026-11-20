import { Routes } from '@angular/router';
import { authGuard } from './auth.guard';
import { ClaimList } from './claim-list/claim-list';

// หมายเหตุ: 'claims/new' ต้องอยู่ก่อน 'claims/:id' ไม่งั้น "new" จะถูกอ่านเป็น id
export const routes: Routes = [
  { path: '', redirectTo: 'claims', pathMatch: 'full' },
  { path: 'claims', component: ClaimList },
  {
    path: 'claims/new',
    canActivate: [authGuard],
    // lazy loading: โค้ดฟอร์มถูกแยกเป็น chunk โหลดตอนเข้าหน้านี้
    loadComponent: () => import('./claim-form/claim-form').then((m) => m.ClaimForm),
  },
  {
    path: 'claims/:id',
    loadComponent: () => import('./claim-detail/claim-detail').then((m) => m.ClaimDetail),
  },
];
