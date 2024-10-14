// place files you want to import through the `$lib` alias in this folder.
export const API_GATEWAY = "http://localhost:3001"
export const FRONTEND_URL = "http://localhost:5173"
export  interface UrlEntry {
    id: string;
    originalUrl: string;
    shortUrl: string;
    createdAt: string;
    clickCount: number;
}

