import { API_GATEWAY } from '$lib';
import { json } from '@sveltejs/kit';
import ky from 'ky';

export async function POST({ request }) {
    try {
        const { userId } = await request.json();

        if (!userId) {
            return json({
                error: 'User ID is required'
            }, { status: 400 });
        }
        
        ky.post(`${API_GATEWAY}/users`, {
            json: {
                userId: userId,
            },
        }).catch((error) => {
            console.error("Failed to register new user:", error);
        });
        
        return json({
            success: true,
            message: 'User registered successfully',
            userId
        }, { status: 201 });

    } catch (error) {
        console.error('Error registering user:', error);
        
        return json({
            error: 'Internal server error'
        }, { status: 500 });
    }
}