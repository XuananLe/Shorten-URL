// place files you want to import through the `$lib` alias in this folder.
export const API_GATEWAY = "http://localhost:3000"
export const FRONTEND_URL = "https://integer-wild-complexity-printers.trycloudflare.com"
export  interface UrlEntry {
    id: string;
    originalUrl: string;
    shortUrl: string;
    createdAt: string;
    clickCount: number;
}

