// Mock API สำหรับรันหน้า Day 19 — ไม่ใช่ Day 8 backend จริง แค่จำลอง contract ให้ใกล้เคียง
//   GET  /claims?status&limit&offset -> { items, limit, offset }
//   GET  /claims/{id}                -> claim | 404 {error}
//   POST /claims  (header Idempotency-Key บังคับ, body {patient_ref, amount}) -> 201 | 200 (replay) | 400 | 409
// กฎ validate คัดลอกจาก day08-backend/internal/domain/claim.go (patient_ref FAKE-[A-Z0-9]{3,12}, amount 1..1e9 สตางค์,
// ห้ามมีฟิลด์อื่น) ต่างจากของจริงที่ id เป็นเลขลำดับ ไม่ใช่ uuid และข้อมูลอยู่ในหน่วยความจำ (หายเมื่อปิด server)
// ไม่ใส่ CORS header โดยตั้งใจ เพื่อให้เห็นว่าต้องผ่าน dev proxy (/api) ถึงจะเรียกได้
// ใช้: node mock-api/server.mjs   |  สลับ fail:  curl -X POST localhost:8080/__fail   (เรียกอีกครั้งเพื่อปิด)
import http from 'node:http';

const now = '2026-11-01T09:00:00Z';
const claims = [
  { id: 'CLM-001', patient_ref: 'FAKE-AB12', amount: 150000, status: 'PENDING' },
  { id: 'CLM-002', patient_ref: 'FAKE-CD34', amount: 3200000, status: 'APPROVED' },
  { id: 'CLM-003', patient_ref: 'FAKE-EF56', amount: 480000, status: 'REJECTED', decision_reason: 'amount_over_limit' },
  { id: 'CLM-004', patient_ref: 'FAKE-GH78', amount: 76000, status: 'PENDING' },
  { id: 'CLM-005', patient_ref: 'FAKE-JK90', amount: 1250000, status: 'APPROVED' },
].map((c) => ({ ...c, created_at: now, updated_at: now }));

const byKey = new Map(); // Idempotency-Key -> { hash, claim }
const DELAY_MS = Number(process.env.DELAY_MS ?? 800);
const MAX_AMOUNT = 1_000_000_000;
let failing = false;

const send = (res, code, body, headers = {}) => {
  res.writeHead(code, { 'Content-Type': 'application/json', ...headers });
  res.end(JSON.stringify(body));
};

http.createServer((req, res) => {
  const url = new URL(req.url, 'http://x');
  if (req.method === 'POST' && url.pathname === '/__fail') {
    failing = !failing;
    res.end(`failing=${failing}\n`);
    return;
  }
  const detail = url.pathname.match(/^\/claims\/([^/]+)$/);

  const handle = async () => {
    if (failing) { res.writeHead(503); res.end('unavailable'); return; }

    if (req.method === 'GET' && url.pathname === '/claims') {
      const status = url.searchParams.get('status');
      const limit = Number(url.searchParams.get('limit') ?? 20);
      const offset = Number(url.searchParams.get('offset') ?? 0);
      const items = claims.filter((c) => !status || c.status === status).slice(offset, offset + limit);
      return send(res, 200, { items, limit, offset });
    }

    if (req.method === 'GET' && detail) {
      const c = claims.find((x) => x.id === decodeURIComponent(detail[1]));
      return c ? send(res, 200, c) : send(res, 404, { error: 'claim not found' });
    }

    if (req.method === 'POST' && url.pathname === '/claims') {
      const key = req.headers['idempotency-key'];
      if (!key) return send(res, 400, { error: 'Idempotency-Key header is required' });
      let raw = '';
      for await (const chunk of req) raw += chunk;
      let body;
      try { body = JSON.parse(raw); } catch { body = null; }
      const keys = body && typeof body === 'object' ? Object.keys(body).sort().join(',') : '';
      if (keys !== 'amount,patient_ref') {
        return send(res, 400, { error: 'request body must be valid JSON with patient_ref and amount' });
      }
      if (!/^FAKE-[A-Z0-9]{3,12}$/.test(body.patient_ref)) {
        return send(res, 400, { error: 'invalid input: patient_ref must match FAKE-[A-Z0-9]{3,12}' });
      }
      if (!Number.isInteger(body.amount) || body.amount <= 0 || body.amount > MAX_AMOUNT) {
        return send(res, 400, { error: `invalid input: amount must be between 1 and ${MAX_AMOUNT}` });
      }
      const hash = `${body.patient_ref}|${body.amount}`;
      const seen = byKey.get(key);
      if (seen) {
        if (seen.hash !== hash) return send(res, 409, { error: 'Idempotency-Key was already used with a different request' });
        return send(res, 200, seen.claim, { 'Idempotent-Replayed': 'true', Location: `/claims/${seen.claim.id}` });
      }
      const id = `CLM-${String(claims.length + 1).padStart(3, '0')}`;
      const claim = { id, patient_ref: body.patient_ref, amount: body.amount, status: 'PENDING', created_at: new Date().toISOString(), updated_at: new Date().toISOString() };
      claims.push(claim);
      byKey.set(key, { hash, claim });
      return send(res, 201, claim, { Location: `/claims/${id}` });
    }

    res.writeHead(404); res.end();
  };

  setTimeout(() => handle().catch(() => { res.writeHead(500); res.end(); }), DELAY_MS);
}).listen(8080, () => console.log('mock api on :8080 (delay', DELAY_MS, 'ms)'));
