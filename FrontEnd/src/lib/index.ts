// place files you want to import through the `$lib` alias in this folder.
export const API_GATEWAY = "http://localhost:3001"
export const FRONTEND_URL = "https://characters-eclipse-ya-forests.trycloudflare.com"
export  interface UrlEntry {
    id: string;
    originalUrl: string;
    shortUrl: string;
    createdAt: string;
    clickCount: number;
}

