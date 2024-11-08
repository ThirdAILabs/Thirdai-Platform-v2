'use client';

import { SessionProvider } from 'next-auth/react';
import { ReactNode } from 'react';
import { Session } from 'next-auth';
import KeycloakIssuerSetter from '@/components/KeycloakIssuerSetter';

interface ProvidersProps {
  children: ReactNode;
  session?: Session; // The session prop should be optional (Session | undefined)
}

export function Providers({ children, session }: ProvidersProps) {
  return (
    <SessionProvider session={session}>
      {/* Add the KeycloakIssuerSetter to dynamically set the kc_issuer cookie */}
      <KeycloakIssuerSetter />
      {children}
    </SessionProvider>
  );
}