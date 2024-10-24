'use client';
import { signIn, useSession } from 'next-auth/react';
import { useEffect } from 'react';

export default function SessionGuard() {
  const { data: session } = useSession();

  console.log("Client Side Access Token Update:", session);

  console.log("Window Type: ", typeof window)

  useEffect(() => {


    console.log("Client Side Access Token Update 2:", session);

    console.log("Window Type 2: ", typeof window)
    if (session?.error === 'RefreshAccessTokenError') {
      signIn('keycloak');
    }

    if (session?.accessToken) {
      localStorage.setItem('accessToken', session.accessToken);
      console.log('Updated localStorage with new accessToken:', session.accessToken);
    }

  }, [session])



  return null;
}
