import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    ramp_rps: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 1000,
      stages: [
        { target: 300, duration: '30s' },
        { target: 600, duration: '30s' },
        { target: 900, duration: '30s' },
        { target: 1200, duration: '30s' },
        { target: 0, duration: '20s' },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<150', 'p(99)<400'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/hello');
  check(res, { 'status 200': (r) => r.status === 200 });
}
