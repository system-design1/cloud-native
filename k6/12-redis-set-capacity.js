import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '15s', target: 100 },
    { duration: '30s', target: 500 },
    { duration: '30s', target: 1000 },
    { duration: '30s', target: 2000 },
    { duration: '30s', target: 3000 },
    { duration: '30s', target: 4000 },
    { duration: '30s', target: 1500 },
    { duration: '20s', target: 0 },
  ],
  thresholds: {
    'http_req_failed': ['rate<0.01'],
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
  },
  discardResponseBodies: true,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TENANT_ID = __ENV.TENANT_ID || '1';
const OTP_CODE = __ENV.OTP_CODE || '123456';

// Generate random phone number in range 09120000000 - 09129999999
function randomPhoneNumber() {
  const randomPart = Math.floor(Math.random() * 10000000)
    .toString()
    .padStart(7, '0');

  return `0912${randomPart}`;
}

export default function () {
  const phoneNumber = randomPhoneNumber();

  const key = `otp:${TENANT_ID}:${phoneNumber}`;

  const value = `{"tenant_id":"${TENANT_ID}","phone_number":"${phoneNumber}","otp_code":"${OTP_CODE}"}`;

  const url = `${BASE_URL}/v1/redis/set?key=${encodeURIComponent(
    key
  )}&value=${encodeURIComponent(value)}`;

  const res = http.post(url, null, {
    timeout: '2s',
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
  });
}
