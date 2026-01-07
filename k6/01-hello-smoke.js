import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = { vus: 1, duration: '10s' };

export default function () {
  const res = http.get('http://localhost:8080/hello');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}


/* 
k6 run hello-smoke.js

k6 run --vus 20 --duration 30s hello-smoke.js


*/

/* 
Installation - method 1:

sudo apt-get update
sudo apt-get install -y gnupg ca-certificates

# اضافه کردن کلید و ریپو
sudo gpg -k
sudo gpg --no-default-keyring \
  --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 \
  --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69

echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" \
  | sudo tee /etc/apt/sources.list.d/k6.list

sudo apt-get update
sudo apt-get install -y k6

==================


Installation - method 2:

sudo apt-get update
sudo apt-get install -y gnupg ca-certificates curl

sudo rm -f /usr/share/keyrings/k6-archive-keyring.gpg
curl -fsSL https://dl.k6.io/key.gpg | gpg --dearmor | sudo tee /usr/share/keyrings/k6-archive-keyring.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" \
  | sudo tee /etc/apt/sources.list.d/k6.list

sudo apt-get update
sudo apt-get install -y k6

=================

Installation - method 3:

sudo apt update
sudo apt install -y snapd
sudo snap install k6

*/