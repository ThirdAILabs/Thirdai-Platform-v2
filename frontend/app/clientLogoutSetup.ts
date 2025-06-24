'use client';

import { useContext, useEffect } from 'react';
import { setLogoutFunction } from '../lib/backend';
import { UserContext } from './user_wrapper';

export default function ClientLogoutSetup() {
  const { logout } = useContext(UserContext);

  useEffect(() => {
    setLogoutFunction(logout);
  }, [logout]);

  return null;
}
