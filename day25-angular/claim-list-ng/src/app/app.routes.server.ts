import { RenderMode, ServerRoute } from '@angular/ssr';

// วันที่ 9: เลือก render mode ราย route
// - Server: เรนเดอร์ตอนมี request (ข้อมูลเคลมเปลี่ยนตลอด จึงไม่ prerender ตอน build)
// - Client: หน้าฟอร์มต้องล็อกอิน (สถานะล็อกอินเป็นของ browser แต่ละเครื่อง server ไม่รู้) ส่งแค่ shell
export const serverRoutes: ServerRoute[] = [
  { path: '', renderMode: RenderMode.Server },
  { path: 'claims', renderMode: RenderMode.Server },
  { path: 'claims/new', renderMode: RenderMode.Client },
  { path: 'claims/new-signal', renderMode: RenderMode.Client },
  { path: 'claims/:id', renderMode: RenderMode.Server },
  { path: '**', renderMode: RenderMode.Server },
];
