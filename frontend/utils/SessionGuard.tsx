// utils/SessionGuard.tsx
'use client';

import { useSession } from 'next-auth/react';
import { useEffect, useState } from 'react';
import { signIn } from 'next-auth/react';
import { useRouter } from 'next/navigation';

export default function SessionGuard() {
  const router = useRouter();
  const { data: session } = useSession();
  const [tokenCheckInterval, setTokenCheckInterval] = useState<NodeJS.Timeout | null>(null);

  const checkAndRefreshToken = async () => {
    const accessToken = session?.token.accessToken;
    const expiresAt = session?.expiresAt || 0;

    if (accessToken && Date.now() >= expiresAt * 1000 - 60 * 1000) {
      // Token is about to expire, attempt to refresh
      router.push('/api/auth/login');
    }
  };

  useEffect(() => {
    if (session?.error === 'RefreshAccessTokenError') {
      // Token refresh failed, redirect to sign-in page
      signIn('keycloak');
    }

    if (session?.token.accessToken) {
      // Store access token in localStorage if needed
      localStorage.setItem('accessToken', session?.token.accessToken);

      if (!tokenCheckInterval) {
        // Set up interval to check token expiration
        const interval = setInterval(checkAndRefreshToken, 60 * 1000);
        setTokenCheckInterval(interval);
      }
    }

    return () => {
      if (tokenCheckInterval) {
        clearInterval(tokenCheckInterval);
        setTokenCheckInterval(null);
      }
    };
  }, [session, tokenCheckInterval]);

  return null;
}
