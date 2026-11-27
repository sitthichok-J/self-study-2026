import {
  AngularNodeAppEngine,
  createNodeRequestHandler,
  isMainModule,
  writeResponseToNodeResponse,
} from '@angular/ssr/node';
import express from 'express';
import { join } from 'node:path';
import { createApiProxy } from './api-proxy';

const browserDistFolder = join(import.meta.dirname, '../browser');

const app = express();
const angularApp = new AngularNodeAppEngine();

/**
 * วันที่ 9: proxy /api -> backend (เหมือน proxy.conf.json ของ ng serve)
 * ตอน SSR, HttpClient ฝั่ง server เรียก URL สัมพัทธ์ '/api/claims' เข้า origin ของ request เอง
 * จึงต้องมีทางส่งต่อไป backend ที่นี่ ตั้งปลายทางด้วย BACKEND_URL (ค่าเริ่มต้น http://localhost:8080)
 */
const backendUrl = process.env['BACKEND_URL'] ?? 'http://localhost:8080';
app.use('/api', createApiProxy(backendUrl));

/**
 * Serve static files from /browser
 */
app.use(
  express.static(browserDistFolder, {
    maxAge: '1y',
    index: false,
    redirect: false,
  }),
);

/**
 * Handle all other requests by rendering the Angular application.
 */
app.use((req, res, next) => {
  angularApp
    .handle(req)
    .then((response) => (response ? writeResponseToNodeResponse(response, res) : next()))
    .catch(next);
});

/**
 * Start the server if this module is the main entry point, or it is ran via PM2.
 * The server listens on the port defined by the `PORT` environment variable, or defaults to 4000.
 */
if (isMainModule(import.meta.url) || process.env['pm_id']) {
  const port = process.env['PORT'] || 4000;
  app.listen(port, (error) => {
    if (error) {
      throw error;
    }

    console.log(`Node Express server listening on http://localhost:${port}`);
  });
}

/**
 * Request handler used by the Angular CLI (for dev-server and during build) or Firebase Cloud Functions.
 */
export const reqHandler = createNodeRequestHandler(app);
