import http from 'k6/http';
import { check } from 'k6';

/**
 * Tenant Settings Lookup — Constant Arrival Rate (steady RPS)
 *
 * Purpose:
 * - Validate the service can SUSTAIN a target RPS while calling:
 *     GET {BASE_URL}/v1/otp/tenant-settings/:id
 *
 * Run examples:
 *   k6 run 10-tenant-settings-constant-arrival-rate.js
 *   BASE_URL=http://localhost:8080 RATE=10000 DURATION=2m k6 run 10-tenant-settings-constant-arrival-rate.js
 *   MIN_ID=1 MAX_ID=20000 RATE=12000 k6 run 10-tenant-settings-constant-arrival-rate.js
 *   TENANT_ID=12345 RATE=15000 k6 run 10-tenant-settings-constant-arrival-rate.js
 */

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Target steady rate (RPS)
const RATE = Number(__ENV.RATE || 10000);
const DURATION = __ENV.DURATION || '2m';

// VU sizing for arrival-rate executor
const PRE_VUS = Number(__ENV.PRE_VUS || 1000);
const MAX_VUS = Number(__ENV.MAX_VUS || 5000);

// Tenant id selection
const TENANT_ID = __ENV.TENANT_ID ? Number(__ENV.TENANT_ID) : null;
const MIN_ID = Number(__ENV.MIN_ID || 1);
const MAX_ID = Number(__ENV.MAX_ID || 20000);

function randomTenantId() {
  const span = (MAX_ID - MIN_ID + 1);
  return MIN_ID + Math.floor(Math.random() * span);
}

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

  // Logical default thresholds for a simple PK lookup on localhost.
  // Tune after you capture a baseline on your machine.
  thresholds: {
    http_req_failed: ['rate<0.01'],

    // For a single indexed SELECT + JSON response, these are reasonable starting SLOs locally.
    http_req_duration: ['p(95)<50', 'p(99)<150'],

    // Key arrival-rate capacity signal:
    // Non-zero means k6 could not keep up with the target rate.
    dropped_iterations: ['count==0'],
  },
};

export default function () {
  const id = TENANT_ID ?? randomTenantId();
  const url = `${BASE_URL}/v1/otp/tenant-settings/${id}`;

  const res = http.get(url, {
    tags: { name: 'GET /v1/otp/tenant-settings/:id' },
    timeout: '2s',
  });

  check(res, { 'status 200': (r) => r.status === 200 });
}
