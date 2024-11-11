'use client';

import { useEffect, useContext } from 'react';
import { userEmailLoginWithAccessToken } from '@/lib/backend';
import { UserContext } from '@/app/user_wrapper';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import Login from '@/components/Login';

export default function ClientHome() {
  const { setAccessToken } = useContext(UserContext);
  const { data: session } = useSession();
  const router = useRouter();

  useEffect(() => {
    const accessToken = session?.token.accessToken;
    console.log("Access Token: ", accessToken)
    if (accessToken) {
      userEmailLoginWithAccessToken(accessToken, setAccessToken)
        .then(() => {
          router.push('/');
        })
        .catch((error) => {
          console.error('Failed to log in with email using the access token:', error);
        });
    }
  }, [session]);

  return (
    <div className="flex flex-col space-y-3 justify-center items-center h-screen">
      {!session && <Login />}
    </div>
  );
}
