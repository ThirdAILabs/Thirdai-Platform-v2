import './globals.css';
import { Analytics } from '@vercel/analytics/react';
import UserWrapper from './user_wrapper';
import SessionGuard from '@/utils/SessionGuard';
import { Providers } from './Providers';
import ThemeProvider from '../theme';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth';
import ClientLogoutSetup from './clientLogoutSetup';

export const metadata = {
  title: 'ThirdAI Platform',
  description: 'Democratize AI for everyone.',
};

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const session = await getServerSession(authOptions);

  return (
    <html lang="en">
      <body className="flex min-h-screen w-full flex-col bg-muted/40">
        <Providers session={session ?? undefined}>
          <ThemeProvider>
            <UserWrapper>
              <ClientLogoutSetup />
              {children}
            </UserWrapper>
          </ThemeProvider>
          <SessionGuard />
        </Providers>
      </body>
      <Analytics />
    </html>
  );
}
