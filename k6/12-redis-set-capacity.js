import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '15s', target: 500 },   // warm-up
    { duration: '30s', target: 1000 },
    { duration: '30s', target: 2000 },
    { duration: '30s', target: 4000 },
    { duration: '30s', target: 5000 },
    { duration: '30s', target: 6000 },  // push hard to find the ceiling
    { duration: '30s', target: 1500 },  // recovery observation
    { duration: '20s', target: 0 },     // cool-down
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200', 'p(99)<500'],
  },
  discardResponseBodies: true,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TENANT_ID = __ENV.TENANT_ID || '1';
const OTP_CODE = __ENV.OTP_CODE || '123456';

// Limit key cardinality to keep Redis memory + benchmark stable.
// You can increase this if you want, but keep it bounded.
const KEY_SPACE = Number(__ENV.KEY_SPACE || 200000); // e.g., 200k unique phones

// Sample checks to reduce overhead at very high RPS
const CHECK_SAMPLE_RATE = Number(__ENV.CHECK_SAMPLE_RATE || 0.01); // 1%

// Generate a deterministic-ish phone number in range 09120000000 - 09129999999
// but bounded by KEY_SPACE to avoid unbounded growth.
function phoneFromIndex(i) {
  // i in [0, 9_999_999], last 7 digits
  const last7 = (i % 10000000).toString().padStart(7, '0');
  return `0912${last7}`;
}

export default function () {
  // Deterministic distribution across VUs/iterations (no heavy randomness needed)
  const idx = (__VU * 100000 + __ITER) % KEY_SPACE;
  const phoneNumber = phoneFromIndex(idx);

  const key = `otp:${TENANT_ID}:${phoneNumber}`;

  // Keep payload small & stable
  const value = `{"tenant_id":"${TENANT_ID}","phone_number":"${phoneNumber}","otp_code":"${OTP_CODE}"}`;

  const url =
    `${BASE_URL}/v1/redis/set` +
    `?key=${encodeURIComponent(key)}` +
    `&value=${encodeURIComponent(value)}`;

  const res = http.post(url, null, {
    timeout: '2s',
    // IMPORTANT: prevent high-cardinality metrics by forcing a stable request name
    tags: { name: 'POST /v1/redis/set' },
  });

  // Only sample checks to reduce overhead under high load
  if (Math.random() < CHECK_SAMPLE_RATE) {
    check(res, {
      'status is 200': (r) => r.status === 200,
    });
  }
}
