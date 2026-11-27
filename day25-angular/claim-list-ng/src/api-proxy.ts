import type { Request, Response } from 'express';

// วันที่ 9: ส่งต่อ /api/* ไป backend (เหมือน proxy.conf.json ของ ng serve) ใช้ตอน SSR
// แยกออกจาก server.ts เพื่อให้เทสต์ได้โดยไม่ต้องเปิด server จริง
export const FORWARD_HEADERS = ['content-type', 'accept', 'idempotency-key'];

export function createApiProxy(backendUrl: string, fetchFn: typeof fetch = fetch) {
  return async (req: Request, res: Response): Promise<void> => {
    try {
      const headers: Record<string, string> = {};
      for (const name of FORWARD_HEADERS) {
        const value = req.headers[name];
        if (typeof value === 'string') headers[name] = value;
      }
      const hasBody = req.method !== 'GET' && req.method !== 'HEAD';
      const upstream = await fetchFn(backendUrl + req.url, {
        method: req.method,
        headers,
        body: hasBody ? (req as unknown as BodyInit) : undefined,
        ...(hasBody ? { duplex: 'half' } : {}),
      } as RequestInit);
      res.status(upstream.status);
      const type = upstream.headers.get('content-type');
      if (type) res.setHeader('content-type', type);
      res.send(Buffer.from(await upstream.arrayBuffer()));
    } catch {
      res.status(502).json({ error: 'backend_unreachable' });
    }
  };
}
