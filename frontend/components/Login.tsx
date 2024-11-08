'use client';
import { signIn } from 'next-auth/react';
import { useState } from 'react';

export default function Login() {
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    console.log('[Login] Login button clicked.');
    setLoading(true);
    try {
      console.log('[Login] Initiating Keycloak sign-in...');
      await signIn('keycloak');
      console.log('[Login] Sign-in process completed.');
    } catch (error) {
      console.error('[Login] Login process failed:', error);
    } finally {
      setLoading(false);
      console.log('[Login] Resetting loading state.');
    }
  };

  return (
    <div className="flex flex-col items-center space-y-4">
      <button
        onClick={handleLogin}
        className="bg-sky-500 hover:bg-sky-700 px-5 py-2 text-sm leading-5 rounded-full font-semibold text-white"
        disabled={loading}
      >
        {loading ? 'Signing in...' : 'Sign in with Keycloak'}
      </button>
      {loading && (
        <p className="text-gray-500 text-sm">Please wait while we redirect you to Keycloak...</p>
      )}
    </div>
  );
}
