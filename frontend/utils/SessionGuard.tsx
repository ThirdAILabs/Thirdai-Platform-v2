'use client';
import { signIn, useSession } from 'next-auth/react';
import { useEffect, useState } from 'react';

export default function SessionGuard() {
  const { data: session } = useSession();
  const [tokenCheckInterval, setTokenCheckInterval] = useState<NodeJS.Timeout | null>(null);

  const checkAndRefreshToken = async () => {
    const accessToken = session?.accessToken;
    const expiresAt = session?.expiresAt || 0;

    console.log('[SessionGuard] Checking token validity...');
    console.log(`[SessionGuard] Access Token: ${accessToken}`);
    console.log(`[SessionGuard] Token Expires At: ${expiresAt}`);
    console.log(`[SessionGuard] Current Time: ${Math.floor(Date.now() / 1000)}`);

    if (accessToken && Date.now() >= expiresAt * 1000 - 60 * 1000) {
      console.log('[SessionGuard] Access token is about to expire. Triggering refresh...');
      try {
        await signIn('keycloak'); // Trigger refresh using NextAuth's built-in refresh mechanism
        console.log('[SessionGuard] Token refreshed successfully.');
      } catch (error) {
        console.error('[SessionGuard] Error refreshing token:', error);
      }
    } else {
      console.log('[SessionGuard] Access token is still valid.');
    }
  };

  useEffect(() => {
    console.log('[SessionGuard] useEffect triggered with session:', session);

    if (session?.error === 'RefreshAccessTokenError') {
      console.warn('[SessionGuard] Session error: RefreshAccessTokenError. Attempting to sign in...');
      signIn('keycloak');
    }

    if (session?.accessToken) {
      console.log('[SessionGuard] Access token available. Storing in localStorage...');
      localStorage.setItem('accessToken', session.accessToken);

      if (!tokenCheckInterval) {
        console.log('[SessionGuard] Setting up interval to check token validity.');
        const interval = setInterval(checkAndRefreshToken, 60 * 1000);
        setTokenCheckInterval(interval);
      }
    } else {
      console.warn('[SessionGuard] No access token available in session.');
    }

    return () => {
      if (tokenCheckInterval) {
        console.log('[SessionGuard] Clearing token check interval.');
        clearInterval(tokenCheckInterval);
        setTokenCheckInterval(null);
      }
    };
  }, [session, tokenCheckInterval]);

  return null;
}
