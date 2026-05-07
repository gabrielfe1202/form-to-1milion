import http from 'k6/http';
import { check, sleep } from 'k6';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';


export const options = {
  stages: [
    { duration: '15s', target: 100 }, // sobe até 100 usuários virtuais
    { duration: '15s', target: 1000 }, // sobe até 1000 usuários virtuais
    { duration: '15s', target: 1500 }, // sobe até 1500 usuários virtuais
    { duration: '30s', target: 2200 }, // chega a 2200 usuários simultâneos
    { duration: '15s', target: 100 },    // desacelera
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% das requisições devem responder em < 500ms
  },
};

export default function () {
  const url = 'http://localhost:80/user';
  const payload = JSON.stringify({
    name: `User_${Math.random().toString(36).substring(2, 10)}`,
    email: `test_${Math.random().toString(36).substring(2, 10)}@example.com`,
    document: `${Math.floor(Math.random() * 1e11).toString().padStart(11, '0')}`,
    phone: `${Math.floor(Math.random() * 1e11).toString().padStart(11, '0')}`,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'status is 202': (r) => r.status === 202,
  });

  sleep(0.1);
}

export function handleSummary(data) {
  const now = new Date();
  const timestamp = now.toISOString().replace(/[:.]/g, '-');
  const outputDir = './tests/load/reports';
  const fileName = `${outputDir}/results-${timestamp}.html`;

  console.log(`k6 handleSummary: salvando HTML em ${fileName}`);
  console.log('k6 handleSummary: métricas principais:');
  console.log(`  - requests: ${data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 'N/A'}`);
  console.log(`  - duration p(95): ${data.metrics.http_req_duration ? data.metrics.http_req_duration.values['p(95)'] : 'N/A'}`);
  console.log(`  - erros: ${data.metrics.http_req_failed ? data.metrics.http_req_failed.values.rate * data.metrics.http_reqs.values.count : 0}`);
  console.log(`  - avg duration: ${data.metrics.http_req_duration.values.avg}`);
  console.log(`  - min duration: ${data.metrics.http_req_duration.values.min}`);
  console.log(`  - max duration: ${data.metrics.http_req_duration.values.max}`);
  console.log(`  - taxa de erro: ${data.metrics.http_req_failed.values.rate}`);

  return {
    [fileName]: htmlReport(data),
    stdout: JSON.stringify({
      summary: {
        vus: data.metrics.vus ? data.metrics.vus.value : undefined,
        vus_max: data.metrics.vus_max ? data.metrics.vus_max.value : undefined,
        iterations: data.metrics.iterations ? data.metrics.iterations.count : undefined,
      },
    }, null, 2),
  };
}