import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 100 },
    { duration: '20s',  target: 300 },
    { duration: '20s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200'], // مثال: 95% زیر 200ms
  },
};

export default function () {
  const res = http.get('http://localhost:8080/hello');
  check(res, { '200': (r) => r.status === 200 });
}


/* 

*/