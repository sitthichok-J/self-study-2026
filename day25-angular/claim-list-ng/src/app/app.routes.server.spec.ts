import { RenderMode } from '@angular/ssr';
import { routes } from './app.routes';
import { serverRoutes } from './app.routes.server';

const MODE = { [RenderMode.Server]: 'Server', [RenderMode.Client]: 'Client', [RenderMode.Prerender]: 'Prerender' } as Record<number, string>;

describe('server routes (SSR)', () => {
  it('baseline: render mode ราย route', () => {
    expect(serverRoutes.map((r) => `${r.path || '(root)'} → ${MODE[r.renderMode!]}`)).toMatchSnapshot();
  });

  it('ทุก route ที่มี guard ล็อกอินต้องเป็น Client (server ไม่รู้ว่าใครล็อกอิน)', () => {
    const guarded = routes.filter((r) => (r.canActivate ?? []).length > 0).map((r) => r.path!);
    expect(guarded.length).toBeGreaterThan(0);
    for (const path of guarded) {
      expect(serverRoutes.find((r) => r.path === path)?.renderMode).toBe(RenderMode.Client);
    }
  });

  it('ไม่ใช้ Prerender เพราะข้อมูลเคลมมาจาก API สด', () => {
    expect(serverRoutes.some((r) => r.renderMode === RenderMode.Prerender)).toBe(false);
  });

  it('ทุก path ใน routes มี entry ใน serverRoutes (หรือถูกคลุมด้วย **)', () => {
    const declared = serverRoutes.map((r) => r.path);
    for (const r of routes) expect(declared).toContain(r.path);
  });
});
