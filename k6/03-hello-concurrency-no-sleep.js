/**
 * تست بار همزمانی (Concurrency Load Test) برای endpoint /hello
 * 
 * این تست بدون sleep اجرا می‌شود - هر VU تا حد امکان سریع درخواست می‌زند.
 * هدف: پیدا کردن ظرفیت حداکثری سیستم و نقطه شکست.
 * 
 * برای اطلاعات بیشتر، به docs/LOAD_TESTING_K6_HELLO_CONCURRENCY.md مراجعه کنید.
 * 
 * نحوه اجرا:
 *   k6 run k6/03-hello-concurrency-no-sleep.js
 */

import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    // { duration: '10s', target: 50 },   // 10s: افزایش به 50 VU (warm-up)
    // { duration: '20s', target: 100 },  // 20s: افزایش به 100 VU
    // { duration: '20s', target: 200 },  // 20s: افزایش به 200 VU
    // { duration: '20s', target: 300 },  // 20s: افزایش به 300 VU
    // { duration: '10s', target: 0 },     // 10s: کاهش به 0 VU (cooldown)

    // test with higher rate
    { duration: '10s', target: 100 },   
    { duration: '20s', target: 200 },  
    { duration: '20s', target: 400 },  
    { duration: '20s', target: 600 },  
    { duration: '10s', target: 0 },    
  ],
  thresholds: {
    // نرخ خطا باید کمتر از 1% باشد
    http_req_failed: ['rate<0.01'],
    // 95% درخواست‌ها باید زیر 50ms باشند
    // 99% درخواست‌ها باید زیر 150ms باشند
    http_req_duration: ['p(95)<50', 'p(99)<150'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/hello');
  check(res, { 'status 200': (r) => r.status === 200 });
  // توجه: sleep وجود ندارد - هر VU تا حد امکان سریع درخواست می‌زند
}
