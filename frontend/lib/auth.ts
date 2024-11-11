// auth.ts
import NextAuth from 'next-auth';
import KeycloakProvider from 'next-auth/providers/keycloak';
import type { Session, Account, User, Profile  } from 'next-auth';
import { NextRequest } from 'next/server';
import { JWT } from 'next-auth/jwt';
import type { NextAuthConfig } from 'next-auth';
import { cookies } from 'next/headers';

export const { auth, handlers, signIn, signOut } = NextAuth((req?: NextRequest): NextAuthConfig => {
  
  const cookieStore = cookies();
  const hostInfo = cookieStore.get('hostInfo')?.value;

  // Construct the issuer dynamically
  const issuer = `${hostInfo}/keycloak/realms/ThirdAI-Platform`;


  const authConfig = {
    providers: [
      KeycloakProvider({
        clientId: process.env.AUTH_KEYCLOAK_CLIENT_ID!,
        clientSecret: process.env.AUTH_KEYCLOAK_CLIENT_SECRET!,
        issuer: issuer,
      }),
    ],
    secret: process.env.AUTH_SECRET,
    session: {
      strategy: 'jwt' as 'jwt',
      maxAge: 60 * 30, // 30 minutes
    },
    trustHost: true,
    callbacks: {
      async jwt({
        token,
        account,
      }: {
        token: JWT;
        account?: Account | null;
      }) {
        // Only process account when it's provided
        if (account) {
          token.idToken = account.id_token;
          token.accessToken = account.access_token;
          token.refreshToken = account.refresh_token;
          token.expiresAt = account.expires_at;
          token.issuer = process.env.AUTH_KEYCLOAK_ISSUER; // Set issuer dynamically
        }
      
        // Token refresh logic
        if (token.expiresAt && Date.now() < token.expiresAt * 1000 - 60 * 1000) {
          return token;
        }
      
        // Refresh the access token
        try {
          const response = await fetch(`${token.issuer}/protocol/openid-connect/token`, {
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: new URLSearchParams({
              client_id: process.env.AUTH_KEYCLOAK_CLIENT_ID!,
              client_secret: process.env.AUTH_KEYCLOAK_CLIENT_SECRET!,
              grant_type: 'refresh_token',
              refresh_token: token.refreshToken!,
            }),
            method: 'POST',
          });
      
          const tokens = await response.json();
      
          if (!response.ok) throw tokens;
      
          return {
            ...token,
            idToken: tokens.id_token,
            accessToken: tokens.access_token,
            expiresAt: Math.floor(Date.now() / 1000 + tokens.expires_in),
            refreshToken: tokens.refresh_token ?? token.refreshToken,
          };
        } catch (error) {
          console.error('Error refreshing access token', error);
          return { ...token, error: 'RefreshAccessTokenError' as const };
        }
      },
      async session({ session, token }: { session: Session; token: JWT }) {
        session.token = token;
        session.error = token.error;
        session.expiresAt = token.expiresAt
        return session;
      },
    },
    debug: true,
  };

  return authConfig;
});
