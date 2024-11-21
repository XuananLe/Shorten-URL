import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '1s', target: 5 }, 
        { duration: '1s', target: 5}, 
        { duration: '1s', target: 0 },
    ],
};

const baseUrl = 'http://localhost:3001';
const testUrl = `${baseUrl}/create?url=https://www.google.com&userId=cb9f7d80-691c-4b33-88c6-b1c99dac8cbc`; // Write endpoint
const shortUrl = `${baseUrl}/short/CCaICRin`; // Read endpoint

let isWritePhase = true;

export default function () {
    if (__VU <= options.stages[0].target * 0.2) {
        const createResponse = http.post(testUrl);
        check(createResponse, {
            'status is 201': (r) => r.status === 201,
        });
    } else {
        const isRead = Math.random() < 0.8; // 80% chance to read, 20% to write

        if (isRead) {
            const shortResponse = http.get(shortUrl);
            check(shortResponse, {
                'status is 200': (r) => r.status === 200,
            });
        } else {
            const createResponse = http.post(testUrl);
            check(createResponse, {
                'status is 201': (r) => r.status === 201,
            });
        }
    }

    sleep(1);
}
