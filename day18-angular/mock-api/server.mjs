// Mock API สำหรับรันหน้า Day 18 — ไม่ใช่ Day 8 backend จริง แค่จำลอง contract เดียวกัน
// GET /claims?status=&limit=&offset= -> { items, limit, offset }
// ไม่ใส่ CORS header โดยตั้งใจ เพื่อให้เห็นว่าต้องผ่าน dev proxy (/api) ถึงจะเรียกได้
// ใช้: node mock-api/server.mjs   |  สลับ fail:  curl -X POST localhost:8080/__fail   (เรียกอีกครั้งเพื่อปิด)
import http from 'node:http';

const ITEMS = [
  { id: 'CLM-001', patient_ref: 'FAKE-AB12', amount: 150000, status: 'PENDING' },
  { id: 'CLM-002', patient_ref: 'FAKE-CD34', amount: 3200000, status: 'APPROVED' },
  { id: 'CLM-003', patient_ref: 'FAKE-EF56', amount: 480000, status: 'REJECTED', decision_reason: 'amount_over_limit' },
  { id: 'CLM-004', patient_ref: 'FAKE-GH78', amount: 76000, status: 'PENDING' },
  { id: 'CLM-005', patient_ref: 'FAKE-JK90', amount: 1250000, status: 'APPROVED' },
].map((c) => ({ ...c, created_at: '2026-11-01T09:00:00Z', updated_at: '2026-11-01T09:00:00Z' }));

const DELAY_MS = Number(process.env.DELAY_MS ?? 1500);
let failing = false;

http.createServer((req, res) => {
  const url = new URL(req.url, 'http://x');
  if (req.method === 'POST' && url.pathname === '/__fail') {
    failing = !failing;
    res.end(`failing=${failing}\n`);
    return;
  }
  if (req.method === 'GET' && url.pathname === '/claims') {
    setTimeout(() => {
      if (failing) { res.writeHead(503); res.end('unavailable'); return; }
      const status = url.searchParams.get('status');
      const limit = Number(url.searchParams.get('limit') ?? 50);
      const offset = Number(url.searchParams.get('offset') ?? 0);
      const items = ITEMS.filter((c) => !status || c.status === status).slice(offset, offset + limit);
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ items, limit, offset }));
    }, DELAY_MS);
    return;
  }
  res.writeHead(404); res.end();
}).listen(8080, () => console.log('mock api on :8080 (delay', DELAY_MS, 'ms)'));
