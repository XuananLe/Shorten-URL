import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { toast } from "svelte-sonner";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export async function copyToClipboard(shortUrl: string): Promise<void> {
	await navigator.clipboard.writeText(shortUrl);
	toast.success("Copied to clipboard!");
}

