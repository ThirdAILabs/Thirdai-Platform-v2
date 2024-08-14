'use server';

import { deleteModelById } from '@/lib/db';
import { revalidatePath } from 'next/cache';

export async function deleteModel(formData: FormData) {
  // let id = Number(formData.get('id'));
  // await deleteProductById(id);
  // revalidatePath('/');
}
