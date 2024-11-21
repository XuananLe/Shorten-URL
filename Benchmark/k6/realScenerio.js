import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Counter, Rate } from 'k6/metrics';

const createUrlTrend = new Trend('create_url_duration');
const accessUrlTrend = new Trend('access_url_duration');
const failedRequests = new Counter('failed_requests');
const successRate = new Rate('success_rate');

export const options = {
  stages: [
    { duration: '5s', target: 1 }, 
    { duration: '10s', target: 5000 },
    { duration: '5s', target: 0 }, 
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    'create_url_duration': ['p(95)<600'],
    'access_url_duration': ['p(95)<400'],
    'success_rate': ['rate>0.95'],
  },
};

const baseUrl = 'http://localhost:3001';
const createEndpoint = `${baseUrl}/create?url=https://kubernetes.io/docs/concepts/overview/components/&userId=1c8be2ab-694d-40a1-acda-6d2ff09e8b76`;
let shortUrl = '';

export default function () {
  group('Shorten URL creation', () => {
    if (__VU === 1) {
      const startTime = new Date();
      const createRes = http.post(createEndpoint);
      const duration = new Date() - startTime;
      
      createUrlTrend.add(duration);
      
      const success = check(createRes, {
        'Create: status is 201': (r) => r.status === 201,
        'Create: response contains shortUrl': (r) => JSON.parse(r.body).shortUrl !== undefined
      });

      if (!success) {
        failedRequests.add(1);
      }
      successRate.add(success);

      shortUrl = JSON.parse(createRes.body).shortUrl;
    }
  });

  sleep(5);

  group('Access shortened URL', () => {
    if (shortUrl) {
      const shortEndpoint = `${baseUrl}/short/${shortUrl}`;
      
      const startTime = new Date();
      const shortRes = http.get(shortEndpoint);
      const duration = new Date() - startTime;
      
      accessUrlTrend.add(duration);

      const success = check(shortRes, {
        'Short: status is 200': (r) => r.status === 200,
        'Short: response contains originalUrl': (r) => JSON.parse(shortRes.body).originalUrl === 'https://kubernetes.io/docs/concepts/overview/components/',
      });

      if (!success) {
        failedRequests.add(1);
      }
      successRate.add(success);
    }
  });

  sleep(1);
}