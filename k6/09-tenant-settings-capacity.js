import http from 'k6/http';
import { check } from 'k6';

/**
 * Tenant Settings Lookup — Capacity (Ramp-to-Break) — VU-based
 *
 * Purpose:
 * - Push the system until it saturates (RPS plateaus and/or latency rises) while calling:
 *     GET {BASE_URL}/v1/otp/tenant-settings/:id
 *
 * Notes:
 * - VU-based => each VU loops as fast as possible (no sleep).
 * - Use MIN_ID/MAX_ID (or TENANT_ID) to control which tenant(s) are queried.
 */

export const options = {
  stages: [
    { duration: '15s', target: 100 },   // warm-up
    { duration: '30s', target: 500 },
    { duration: '30s', target: 1000 },
    { duration: '30s', target: 2000 },
    { duration: '30s', target: 3000 },
    { duration: '30s', target: 4000 },  // push hard to find the ceiling
    { duration: '30s', target: 1500 },  // recovery observation
    { duration: '20s', target: 0 },     // cool-down
  ],
  // Permissive thresholds for discovery (don't fail too early).
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TENANT_ID = __ENV.TENANT_ID ? Number(__ENV.TENANT_ID) : null;
const MIN_ID = Number(__ENV.MIN_ID || 1);
const MAX_ID = Number(__ENV.MAX_ID || 20000);

function randomTenantId() {
  // Inclusive range [MIN_ID, MAX_ID]
  const span = (MAX_ID - MIN_ID + 1);
  return MIN_ID + Math.floor(Math.random() * span);
}

export default function () {
  const id = TENANT_ID ?? randomTenantId();
  const url = `${BASE_URL}/v1/otp/tenant-settings/${id}`;

  const res = http.get(url, {
    tags: { name: 'GET /v1/otp/tenant-settings/:id' },
    timeout: '2s',
  });

  // Minimal checks to reduce client overhead for capacity discovery.
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
