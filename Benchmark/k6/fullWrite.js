import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '1s', target: 5 }, 
        { duration: '1s', target: 5}, 
        { duration: '1s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
    }
};
const baseUrl = 'http://localhost:3001';
const testUrl = `${baseUrl}/create?url=https://www.google.com&userId=cb9f7d80-691c-4b33-88c6-b1c99dac8cbc`; // Write endpoint

export default function () {
    const createResponse = http.post(testUrl);
    check(createResponse, {
        'status is 201': (r) => r.status === 201,
    });
    console.logF(createResponse.body.shortUrl);

    sleep(0.5);
}