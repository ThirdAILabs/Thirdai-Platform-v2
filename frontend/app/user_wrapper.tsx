'use client';

import { getAccessToken, User, accessTokenUser } from '@/lib/backend';
import { useRouter } from 'next/navigation';
import { useEffect, useState, createContext, SetStateAction, Dispatch, useCallback } from 'react';
import federatedLogout from '@/utils/federatedLogout';

interface UserContext {
  user: User | null;
  accessToken?: string | null;
  setAccessToken: Dispatch<SetStateAction<string | null | undefined>>;
  logout: () => void;
}

// Dummy default values. They will be correctly initialized by the time the context is
// used by the provider's children.
export const UserContext = createContext<UserContext>({
  user: null,
  accessToken: null,
  setAccessToken: (user) => {},
  logout: () => {},
});

export default function UserWrapper({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [accessToken, setAccessToken] = useState<string | null | undefined>(() => {
    // Initialize from localStorage during component mount
    if (typeof window !== 'undefined') {
      return localStorage.getItem('accessToken');
    }
    return null;
  });
  const [user, setUser] = useState<User | null>(null);

  const logout = async () => {
    setAccessToken(null);
    localStorage.removeItem('accessToken');
    setUser(null);
    if (
      process.env.NEXT_PUBLIC_IDENTITY_PROVIDER &&
      process.env.NEXT_PUBLIC_IDENTITY_PROVIDER.toLowerCase().includes('keycloak')
    ) {
      await federatedLogout();
      router.push('/login-keycloak');
    } else {
      router.push('/login-email');
    }
  };

  // Handle access token changes
  useEffect(() => {
    if (!accessToken) {
      setUser(null);
      return;
    }

    const fetchUser = async () => {
      try {
        const thisUser = await accessTokenUser(accessToken);
        if (thisUser) {
          setUser(thisUser);
        } else {
          await logout();
        }
      } catch (error) {
        await logout();
      }
    };

    fetchUser();
  }, [accessToken]);

  useEffect(() => {
    if (accessToken) {
      localStorage.setItem('accessToken', accessToken);
    } else {
      localStorage.removeItem('accessToken');
    }
  }, [accessToken]);
  return (
    <UserContext.Provider value={{ user, accessToken, setAccessToken, logout }}>
      {children}
    </UserContext.Provider>
  );
}
