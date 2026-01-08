import http from 'k6/http';
import { check } from 'k6';

/**
 * Constant Arrival Rate (steady RPS) test for OTP generation endpoint.
 *
 * Target endpoint:
 *   POST {BASE_URL}/v1/otp/code
 *
 * Run examples:
 *   k6 run 08-otp-constant-arrival-rate.js
 *   BASE_URL=http://localhost:8080 k6 run 08-otp-constant-arrival-rate.js
 *   RATE=16000 DURATION=3m k6 run 08-otp-constant-arrival-rate.js
 */

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const RATE = Number(__ENV.RATE || 16000);          // target RPS (tune per your box)
const DURATION = __ENV.DURATION || '2m';           // steady state duration
const PRE_VUS = Number(__ENV.PRE_VUS || 1000);     // pre-allocated VUs
const MAX_VUS = Number(__ENV.MAX_VUS || 3000);     // max VUs k6 may scale up to

export const options = {
  scenarios: {
    steady_rps: {
      executor: 'constant-arrival-rate',
      rate: RATE,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_VUS,
      maxVUs: MAX_VUS,
    },
  },

  thresholds: {
    // Failures should be near-zero on localhost for a simple endpoint.
    http_req_failed: ['rate<0.01'],

    // Slightly looser than /hello because OTP uses crypto/rand.
    http_req_duration: ['p(95)<120', 'p(99)<300'],

    // Key capacity signal for arrival-rate tests:
    // If non-zero => k6 could not keep up with the target rate.
    dropped_iterations: ['count==0'],
  },
};

export default function () {
  const url = `${BASE_URL}/v1/otp/code`;

  const res = http.post(url, null, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'POST /v1/otp/code' },
    timeout: '2s',
  });

  check(res, { 'status 200': (r) => r.status === 200 });
}
