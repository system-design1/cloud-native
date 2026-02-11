import http from 'k6/http';
import { check } from 'k6';

/**
 * Tenant Settings INSERT — Capacity (Ramp-to-Break) — VU-based
 *
 * Purpose:
 * - Push the system until it saturates (RPS plateaus and/or latency rises) while calling:
 *     POST {BASE_URL}/v1/otp/tenant-settings-insert-benchmark
 *
 * Notes:
 * - VU-based => each VU loops as fast as possible (no sleep).
 * - Endpoint is designed for raw INSERT throughput benchmarking (minimal payload).
 */

export const options = {
  stages: [
    { duration: '15s', target: 100 },   // warm-up
    { duration: '30s', target: 500 },
    { duration: '30s', target: 2000 },
    { duration: '30s', target: 3000 },
    { duration: '30s', target: 4000 },
    { duration: '30s', target: 5000 },  // push hard to find the ceiling
    { duration: '30s', target: 3000 },  // recovery observation
    { duration: '20s', target: 0 },     // cool-down
  ],
  // Permissive thresholds for discovery (don't fail too early).
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const URL = `${BASE_URL}/v1/otp/tenant-settings-insert-benchmark`;

export default function () {
  // No payload to keep client overhead minimal (server uses correlation_id for uniqueness).
  const res = http.post(URL, null, {
    tags: { name: 'POST /v1/otp/tenant-settings-insert-benchmark' },
    timeout: '3s',
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
