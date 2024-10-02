// src/app.d.ts
declare global {
    namespace App {
        interface UrlEntry {
            id: string;
            originalUrl: string;
            shortUrl: string;
            createdAt: string;
            clickCount: number;
        }
    }
}

export {};