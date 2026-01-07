import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    steady_rps: {
      executor: 'constant-arrival-rate',
      rate: 500,          // هدف: 500 RPS
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 50,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<100', 'p(99)<250'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/hello');
  check(res, { 'status 200': (r) => r.status === 200 });
}
