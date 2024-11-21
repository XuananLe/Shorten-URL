import { check } from 'k6';
import http from 'k6/http';
export const options = {
    stages: [
        { duration: '1s', target: 5 },
        { duration: '1s', target: 5 },
        { duration: '1s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
    },
};

const baseUrl = 'http://localhost:3001';
const shortUrl = `${baseUrl}/short/CCaICRddd`;

export default function () {
    const shortResponse = http.get(shortUrl);
    check(shortResponse, {
        'status is 404': (r) => r.status === 404,
    });
}