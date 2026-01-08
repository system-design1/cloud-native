import http from 'k6/http';
import { check } from 'k6';

/**
 * OTP Capacity (Ramp-to-Break) — VU-based stress/capacity test
 *
 * Purpose:
 * - Push the system until it clearly saturates (plateau in RPS + rising latency) and/or starts failing.
 *
 * SUT:
 * - POST {BASE_URL}/v1/otp/code
 *
 * Environment (your previous local box):
 * - Ubuntu 24.04, i7-7700 (8 threads), 16GB RAM
 *
 * Notes:
 * - VU-based means each VU runs as fast as possible (no sleep).
 * - Increase the last stage targets if you still don't see a plateau/break.
 * - Keep checks lightweight; heavy checks reduce max throughput.
 */

export const options = {
  stages: [
    // { duration: '15s', target: 10 },    // warm-up
    // { duration: '30s', target: 50 },
    // { duration: '45s', target: 100 },
    // { duration: '60s', target: 200 },
    // { duration: '60s', target: 400 },
    // { duration: '60s', target: 600 },
    // { duration: '60s', target: 800 },
    // { duration: '60s', target: 1000 },  // likely near/above local capacity on this box
    // { duration: '60s', target: 1200 },  // push harder to find breaking point
    // { duration: '20s', target: 0 },     // cool-down


    { duration: '15s', target: 100 },    // warm-up
    { duration: '30s', target: 500 },
    { duration: '30s', target: 1000 },
    { duration: '30s', target: 2000 },
    { duration: '30s', target: 3000 },
    { duration: '30s', target: 5000 },
    { duration: '30s', target: 1500 },  
    { duration: '20s', target: 0 },     // cool-down
  ],

  // Permissive thresholds for capacity discovery (don’t fail too early).
  thresholds: {
    http_req_failed: ['rate<0.05'],              // allow up to 5% while finding the limit
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const url = `${BASE_URL}/v1/otp/code`;

  const res = http.post(url, null, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'POST /v1/otp/code' },
    timeout: '2s', // keeps the run moving when the server stalls
  });

  // Minimal check to avoid client-side overhead
  check(res, { 'status is 200': (r) => r.status === 200 });
}
