import http from 'k6/http';
import { check } from 'k6';

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
  thresholds: {
    'http_req_failed{phase:main}': ['rate<0.01'],
    'http_req_duration{phase:main}': ['p(95)<500', 'p(99)<1000'],
  },
  discardResponseBodies: true,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const TENANT_ID = __ENV.TENANT_ID || '1';
const PHONE_NUMBER = __ENV.PHONE_NUMBER || '09120000000';
const OTP_CODE = __ENV.OTP_CODE || '123456';

const SEED_KEYS = Number(__ENV.SEED_KEYS || 5000);

// Sample checks to reduce overhead at very high RPS
const CHECK_SAMPLE_RATE = Number(__ENV.CHECK_SAMPLE_RATE || 0.01); // 1%

function seedKey(i) {
  return `otp:${TENANT_ID}:${PHONE_NUMBER}:seed:${i}`;
}

export function setup() {
  const value = `{"tenant_id":"${TENANT_ID}","phone_number":"${PHONE_NUMBER}","otp_code":"${OTP_CODE}"}`;

  // Seed keys so GET is always a hit.
  // Use stable "name" tag to avoid high-cardinality metrics due to unique URLs.
  for (let i = 0; i < SEED_KEYS; i++) {
    const key = seedKey(i);
    const url =
      `${BASE_URL}/v1/redis/set` +
      `?key=${encodeURIComponent(key)}` +
      `&value=${encodeURIComponent(value)}`;

    const res = http.post(url, null, {
      timeout: '2s',
      tags: {
        phase: 'setup',
        op: 'redis_seed',
        name: 'POST /v1/redis/set (seed)',
      },
    });

    if (res.status !== 200) {
      throw new Error(`seed failed at i=${i}, status=${res.status}`);
    }
  }

  return { seedKeys: SEED_KEYS };
}

export default function (data) {
  // Choose a key uniformly from the seeded set
  const i = Math.floor(Math.random() * data.seedKeys);
  const key = seedKey(i);

  const url = `${BASE_URL}/v1/redis/get?key=${encodeURIComponent(key)}`;

  const res = http.get(url, {
    timeout: '2s',
    tags: {
      phase: 'main',
      op: 'redis_get',
      // IMPORTANT: stable request name to avoid high-cardinality metrics
      name: 'GET /v1/redis/get',
    },
  });

  // Sample checks to keep overhead low
  if (Math.random() < CHECK_SAMPLE_RATE) {
    check(res, {
      'status is 200': (r) => r.status === 200,
    });
  }
}

/*
# Default seed = 5000
k6 run 13-redis-get-capacity.js

# Different seed size:
SEED_KEYS=20000 k6 run 13-redis-get-capacity.js

# To increase check sampling (debug only; increases overhead):
CHECK_SAMPLE_RATE=1 k6 run 13-redis-get-capacity.js
*/
