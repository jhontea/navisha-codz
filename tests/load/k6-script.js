import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const failureRate = new Rate('failed_requests');
const responseTime = new Trend('response_time');
const problemCount = new Counter('problems_retrieved');

export const options = {
  stages: [
    { duration: '30s', target: 100 },  // Ramp up to 100 VUs over 30s
    { duration: '1m', target: 100 },   // Stay at 100 VUs for 1 minute
    { duration: '30s', target: 0 },    // Ramp down over 30s
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],    // 95% of requests under 2s
    http_req_failed: ['rate<0.01'],       // Less than 1% failure rate
    failed_requests: ['rate<0.05'],       // Custom failure rate under 5%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  group('List Problems', function () {
    // GET /api/problems - list all problems
    const listResp = http.get(`${BASE_URL}/api/problems`, {
      headers: { 'Accept': 'application/json' },
    });

    check(listResp, {
      'list problems - status 200': (r) => r.status === 200,
      'list problems - has data': (r) => r.json('data') !== undefined,
      'list problems - has meta': (r) => r.json('meta') !== undefined,
    });

    responseTime.add(listResp.timings.duration);
    failureRate.add(listResp.status !== 200);

    if (listResp.status === 200) {
      const data = listResp.json('data');
      if (Array.isArray(data)) {
        problemCount.add(data.length);
      }
    }
  });

  group('Get Specific Problem', function () {
    // GET /api/problems/two-sum - get specific problem
    const problemResp = http.get(`${BASE_URL}/api/problems/two-sum`, {
      headers: { 'Accept': 'application/json' },
    });

    check(problemResp, {
      'get problem - status 200': (r) => r.status === 200,
      'get problem - correct ID': (r) => r.json('data.id') === 'two-sum',
      'get problem - no solution leak': (r) => r.json('data.solution') === undefined,
    });

    responseTime.add(problemResp.timings.duration);
    failureRate.add(problemResp.status !== 200);
  });

  group('Get Problem Template', function () {
    // GET /api/problems/two-sum/template
    const templateResp = http.get(`${BASE_URL}/api/problems/two-sum/template`, {
      headers: { 'Accept': 'application/json' },
    });

    check(templateResp, {
      'get template - status 200': (r) => r.status === 200,
      'get template - has template': (r) => r.json('data.template') !== undefined,
    });

    responseTime.add(templateResp.timings.duration);
    failureRate.add(templateResp.status !== 200);
  });

  group('Filter Problems', function () {
    // GET /api/problems?difficulty=easy - filtered list
    const filterResp = http.get(`${BASE_URL}/api/problems?difficulty=easy`, {
      headers: { 'Accept': 'application/json' },
    });

    check(filterResp, {
      'filter problems - status 200': (r) => r.status === 200,
      'filter problems - all easy': (r) => {
        if (r.status !== 200) return false;
        const data = r.json('data');
        return Array.isArray(data) && data.every(p => p.difficulty === 'easy');
      },
    });

    responseTime.add(filterResp.timings.duration);
    failureRate.add(filterResp.status !== 200);
  });

  // Simulate user think-time between requests
  sleep(Math.random() * 2 + 0.5); // 0.5 to 2.5 seconds
}
