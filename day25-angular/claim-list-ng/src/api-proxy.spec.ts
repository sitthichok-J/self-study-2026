import { createApiProxy, FORWARD_HEADERS } from './api-proxy';

function fakeRes() {
  const res: any = { statusCode: 0, headers: {} as Record<string, string>, body: undefined as unknown, json: undefined as unknown };
  res.status = (c: number) => ((res.statusCode = c), res);
  res.setHeader = (k: string, v: string) => ((res.headers[k] = v), res);
  res.send = (b: unknown) => ((res.body = b), res);
  res.json = (b: unknown) => ((res.json = b), res);
  return res;
}

function reply(status: number, body: string, type: string | null = 'application/json') {
  return new Response(type ? body : null, { status, headers: type ? { 'content-type': type } : {} });
}

describe('api proxy (SSR)', () => {
  it('GET: ต่อ URL ท้ายจาก backendUrl + req.url (ตัด /api ออกแล้ว) และส่ง status/body/content-type กลับ', async () => {
    const fetchFn = vi.fn(async () => reply(200, '{"items":[]}'));
    const res = fakeRes();
    await createApiProxy('http://be:8080', fetchFn as any)({ method: 'GET', url: '/claims?limit=11&offset=0', headers: {} } as any, res);
    const [url, init] = fetchFn.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe('http://be:8080/claims?limit=11&offset=0');
    expect(init.method).toBe('GET');
    expect(init.body).toBeUndefined();
    expect(res.statusCode).toBe(200);
    expect(res.headers['content-type']).toBe('application/json');
    expect(Buffer.from(res.body as Buffer).toString()).toBe('{"items":[]}');
  });

  it('ส่งต่อเฉพาะ header ที่อนุญาต (content-type, accept, idempotency-key) ไม่ส่ง cookie/authorization', async () => {
    const fetchFn = vi.fn(async () => reply(201, '{}'));
    const req = {
      method: 'POST',
      url: '/claims',
      headers: { 'content-type': 'application/json', 'idempotency-key': 'k1', cookie: 'a=b', authorization: 'x', accept: 'application/json' },
    };
    await createApiProxy('http://be', fetchFn as any)(req as any, fakeRes());
    const init = (fetchFn.mock.calls[0] as unknown as [string, any])[1];
    expect(init.headers).toEqual({ 'content-type': 'application/json', accept: 'application/json', 'idempotency-key': 'k1' });
    expect(init.body).toBe(req);
    expect(init.duplex).toBe('half');
  });

  it('status ของ backend ผ่านตรง ๆ (เช่น 503 ไม่ถูกแปลง) เพื่อให้ retry interceptor ฝั่ง client ทำงานเหมือนเดิม', async () => {
    const res = fakeRes();
    await createApiProxy('http://be', (async () => reply(503, 'busy', null)) as any)({ method: 'GET', url: '/claims', headers: {} } as any, res);
    expect(res.statusCode).toBe(503);
    expect(res.headers['content-type']).toBeUndefined();
  });

  it('backend ติดต่อไม่ได้ → 502 พร้อม JSON', async () => {
    const res = fakeRes();
    await createApiProxy('http://be', (async () => { throw new Error('ECONNREFUSED'); }) as any)({ method: 'GET', url: '/claims', headers: {} } as any, res);
    expect(res.statusCode).toBe(502);
    expect(res.json).toEqual({ error: 'backend_unreachable' });
  });

  it('baseline: รายการ header ที่ส่งต่อ', () => {
    expect(FORWARD_HEADERS).toMatchSnapshot();
  });
});
