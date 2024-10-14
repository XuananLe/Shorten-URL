import { API_GATEWAY } from '$lib';
import { redirect } from '@sveltejs/kit';
import ky from 'ky';

export async function load({ params }) {
  const id = params.id;

  const data : any = await ky.get(`${API_GATEWAY}/short/${id}`).json();

  throw redirect(307, data.originalUrl);
}