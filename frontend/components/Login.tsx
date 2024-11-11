// components/Login.tsx
'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function Login() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);


  const handleLogin = async () => {
    console.log("Log In called!");
    setLoading(true);
    try {
      router.push('/api/auth/login');
    } catch (error) {
      console.error('Login process failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleLogin}
      className="bg-sky-500 hover:bg-sky-700 px-5 py-2 text-sm leading-5 rounded-full font-semibold text-white"
      disabled={loading}
    >
      {loading ? 'Signing in...' : 'Sign in with Keycloak'}
    </button>
  );
}
