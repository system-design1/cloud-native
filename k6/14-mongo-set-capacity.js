import http from 'k6/http';
import { check, sleep } from 'k6';

// Mongo SET benchmark (mirrors the staged load pattern used in your Redis tests).
// Endpoint: POST /v1/mongo/set?tenant=...&phone=...&value=...&ttl=...

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TENANT   = __ENV.TENANT   || 't1';
const TTL      = __ENV.TTL      || '120s';

// Spread writes across keys to avoid hot-key contention.
const PHONE_PREFIX = __ENV.PHONE_PREFIX || '98912';
const VALUE        = __ENV.VALUE        || '123456';

export const options = {
  stages: [
    { duration: '15s', target: 100 },   // warm-up
    { duration: '30s', target: 500 },
    { duration: '30s', target: 1000 },
    { duration: '30s', target: 2000 },
    { duration: '30s', target: 3000 },
    { duration: '30s', target: 4000 },  // push
    { duration: '30s', target: 1500 },  // recovery observation
    { duration: '20s', target: 0 },     // cool-down
  ],
  thresholds: {
    'http_req_failed{phase:main}': ['rate<0.01'],
    'http_req_duration{phase:main}': ['p(95)<500', 'p(99)<1000'],
  },
  discardResponseBodies: true,
};

function phoneForIteration() {
  // Example: 98912 + 7 digits -> 98912XXXXXXX
  const suffix = String((__VU * 1000000) + (__ITER % 1000000)).padStart(7, '0').slice(0, 7);
  return `${PHONE_PREFIX}${suffix}`;
}

export default function () {
  const phone = phoneForIteration();
  const url =
    `${BASE_URL}/v1/mongo/set` +
    `?tenant=${encodeURIComponent(TENANT)}` +
    `&phone=${encodeURIComponent(phone)}` +
    `&value=${encodeURIComponent(VALUE)}` +
    `&ttl=${encodeURIComponent(TTL)}`;

    const res = http.post(url, null, {
      tags: {
        phase: 'main',
        name: 'POST /v1/mongo/set', 
      },
    });
    

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  // Optional tiny sleep; keep 0 for max throughput
  // sleep(0.001);
}
