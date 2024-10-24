'use client';
import { signIn, useSession } from 'next-auth/react';
import { useEffect, useState } from 'react';

export default function SessionGuard() {
  const { data: session } = useSession();
  const [tokenCheckInterval, setTokenCheckInterval] = useState<NodeJS.Timeout | null>(null);

  const checkAndRefreshToken = async () => {
    const accessToken = session?.accessToken;
    const expiresAt = session?.expiresAt || 0;

    // Check if the token is about to expire (within 1 minute)
    if (accessToken && Date.now() >= expiresAt * 1000 - 60 * 1000) {
      console.log("Access token is about to expire, refreshing token...");
      await signIn('keycloak'); // Trigger refresh using NextAuth's built-in refresh mechanism
    }
  };

  useEffect(() => {
    if (session?.error === 'RefreshAccessTokenError') {
      // If token refresh failed, trigger sign-in again
      signIn('keycloak');
    }

    if (session?.accessToken) {
      // Store access token in localStorage
      localStorage.setItem('accessToken', session.accessToken);
      console.log('Updated localStorage with new accessToken:', session.accessToken);

      // Set up an interval to check the token expiration regularly
      if (!tokenCheckInterval) {
        const interval = setInterval(checkAndRefreshToken, 60 * 1000); // Check every minute
        setTokenCheckInterval(interval);
      }
    }

    return () => {
      // Clear the interval when the component is unmounted
      if (tokenCheckInterval) {
        clearInterval(tokenCheckInterval);
        setTokenCheckInterval(null);
      }
    };
  }, [session, tokenCheckInterval]);

  return null;
}
