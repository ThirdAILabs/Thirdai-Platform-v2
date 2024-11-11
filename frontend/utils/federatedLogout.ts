// utils/federatedLogout.ts
import { signOut } from '@/lib/auth';

export default async function federatedLogout() {
  try {
    const response = await fetch('/federated-logout');
    const data = await response.json();
    if (response.ok) {
      await signOut({ redirect: false });
      window.location.href = data.url;
      return;
    }
    throw new Error(data.error);
  } catch (error) {
    console.error('Error during federated logout:', error);
    await signOut({ redirect: false });
    window.location.href = '/';
  }
}
