import { API_GATEWAY } from '$lib';
import { json } from '@sveltejs/kit';
import ky from 'ky';

export async function POST({ url }) {
    try {
        const originalUrl = url.searchParams.get('url');
        const userId = url.searchParams.get('userId');
        if (!originalUrl || !userId) {
            return json({
                error: 'URL and userId are required'
            }, { status: 400 });
        }
        console.log(originalUrl, userId)
        const data = await ky
            .post(
                `${API_GATEWAY}/create?url=${originalUrl}&userId=${userId}`,
            )
            .json();
        console.log(data);

        return json(data, { status: 200 });  // Properly return the data wrapped in json

    } catch (error) {
        console.error(error);
        return json({
            error: 'Bad Request'
        }, { status: 400 });
    }
}
