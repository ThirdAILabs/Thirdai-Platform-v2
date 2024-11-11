import './globals.css';

import { Analytics } from '@vercel/analytics/react';
import UserWrapper from './user_wrapper';
import SessionGuard from '@/utils/SessionGuard';
import { Providers } from './Providers';

import ThemeProvider from '../theme';
export const metadata = {
  title: 'ThirdAI Platform',
  description: 'Democratize AI for everyone.',
};
import { getServerSession } from 'next-auth';
import { getAuthOptions } from '@/lib/auth';
import { cookies } from 'next/headers';

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const issuerCookie = cookies().get('kc_issuer');
  const issuer = issuerCookie?.value || process.env.KEYCLOAK_ISSUER;

  const authOptions = getAuthOptions(issuer);
  const session = await getServerSession(authOptions);


  return (
    <html lang="en">
      <body className="flex min-h-screen w-full flex-col bg-muted/40">
        <Providers session={session ?? undefined}>
          <ThemeProvider>
            <UserWrapper>{children}</UserWrapper>
          </ThemeProvider>
          <SessionGuard />
        </Providers>
      </body>
      <Analytics />
    </html>
  );
}
