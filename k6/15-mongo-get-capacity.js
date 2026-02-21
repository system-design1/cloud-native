import http from 'k6/http';
import { check } from 'k6';

export const options = {
  setupTimeout: '15m',
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

// Keep naming consistent with your Redis script, but adapted to Mongo endpoints.
const TENANT_ID = __ENV.TENANT_ID || '1';

// For seeding many keys, we generate many phone numbers.
// Default prefix resembles Iranian mobile format; adjust as needed.
const PHONE_PREFIX = __ENV.PHONE_PREFIX || '0912';

const OTP_CODE = __ENV.OTP_CODE || '123456';

const SEED_KEYS = Number(__ENV.SEED_KEYS || 100000);

// Sample checks to reduce overhead at very high RPS
const CHECK_SAMPLE_RATE = Number(__ENV.CHECK_SAMPLE_RATE || 0.01); // 1%

// Mongo SET supports ttl; keep it long enough so keys don't expire during the run.
const TTL = __ENV.TTL || '30m';

// Per-request timeout (not setup timeout). Same idea as your Redis script.
const REQ_TIMEOUT = __ENV.REQ_TIMEOUT || '2s';

function seedPhone(i) {
  // Example: 0912 + 7 digits -> 0912XXXXXXX
  const suffix = String(i).padStart(7, '0').slice(0, 7);
  return `${PHONE_PREFIX}${suffix}`;
}

export function setup() {
  const value = `{"tenant_id":"${TENANT_ID}","phone_number":"${PHONE_PREFIX}0000000","otp_code":"${OTP_CODE}"}`;

  // Seed keys so GET is always a hit.
  // Use stable "name" tag to avoid high-cardinality metrics due to unique URLs.
  for (let i = 0; i < SEED_KEYS; i++) {
    const phone = seedPhone(i);
    const url =
      `${BASE_URL}/v1/mongo/set` +
      `?tenant=${encodeURIComponent(TENANT_ID)}` +
      `&phone=${encodeURIComponent(phone)}` +
      `&value=${encodeURIComponent(value)}` +
      `&ttl=${encodeURIComponent(TTL)}`;

    const res = http.post(url, null, {
      timeout: REQ_TIMEOUT,
      tags: {
        phase: 'setup',
        op: 'mongo_seed',
        name: 'POST /v1/mongo/set (seed)',
      },
    });

    if (res.status !== 200) {
      throw new Error(`seed failed at i=${i}, status=${res.status}`);
    }
  }

  return { seedKeys: SEED_KEYS };
}

export default function (data) {
  // Choose a phone uniformly from the seeded set
  const i = Math.floor(Math.random() * data.seedKeys);
  const phone = seedPhone(i);

  const url =
    `${BASE_URL}/v1/mongo/get` +
    `?tenant=${encodeURIComponent(TENANT_ID)}` +
    `&phone=${encodeURIComponent(phone)}`;

  const res = http.get(url, {
    timeout: REQ_TIMEOUT,
    tags: {
      phase: 'main',
      op: 'mongo_get',
      // IMPORTANT: stable request name to avoid high-cardinality metrics
      name: 'GET /v1/mongo/get',
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
# Default seed = 100000
k6 run 15-mongo-get-capacity.js

# Different seed size:
SEED_KEYS=20000 k6 run 15-mongo-get-capacity.js

# Use a different tenant / phone prefix:
TENANT_ID=1 PHONE_PREFIX=0919 k6 run 15-mongo-get-capacity.js

# Increase/decrease per-request timeout:
REQ_TIMEOUT=1s k6 run 15-mongo-get-capacity.js

# To increase check sampling (debug only; increases overhead):
CHECK_SAMPLE_RATE=1 k6 run 15-mongo-get-capacity.js

# k6 flow
k6 → HTTP → Gin → Mongo Client → MongoDB
*/
