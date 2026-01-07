import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 20 },   // warm-up
    { duration: '5s',  target: 200 },  // spike ناگهانی
    { duration: '20s', target: 200 },  // نگه داشتن spike
    { duration: '10s', target: 20 },   // برگشت
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.02'],
    http_req_duration: ['p(95)<200', 'p(99)<500'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/hello');
  check(res, { 'status 200': (r) => r.status === 200 });
}
