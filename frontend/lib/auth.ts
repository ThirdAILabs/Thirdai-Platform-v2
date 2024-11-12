import { cookies } from 'next/headers';
import { AuthOptions, TokenSet } from 'next-auth';
import { JWT } from 'next-auth/jwt';
import KeycloakProvider from 'next-auth/providers/keycloak';
import type { NextApiRequest, NextApiResponse } from 'next';
import NextAuth from 'next-auth/next';

function requestRefreshOfAccessToken(token: JWT) {
  const issuerCookie = cookies().get('kc_issuer');

  const issuer = token.issuer || issuerCookie?.value || process.env.KEYCLOAK_ISSUER;

  return fetch(`${issuer}/protocol/openid-connect/token`, {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      client_id: process.env.KEYCLOAK_CLIENT_ID,
      client_secret: process.env.KEYCLOAK_CLIENT_SECRET,
      grant_type: 'refresh_token',
      refresh_token: token.refreshToken!,
    }),
    method: 'POST',
    cache: 'no-store',
  });
}


export function getAuthOptions(issuer: string): AuthOptions {

  const authOptions: AuthOptions = {
    providers: [
      KeycloakProvider({
        clientId: process.env.KEYCLOAK_CLIENT_ID,
        clientSecret: process.env.KEYCLOAK_CLIENT_SECRET,
        // Dynamically resolve the issuer using cookies or environment variables
        issuer: issuer,
      }),
    ],
    secret: process.env.NEXTAUTH_SECRET,
    pages: {
      signIn: '/auth/signin',
    },
    session: {
      strategy: 'jwt',
      maxAge: 60 * 30,
    },
    debug: true,
    logger: {
      error(code, ...message) {
        console.error('NextAuth error:', code, message);
      },
      warn(code, ...message) {
        console.warn('NextAuth warning:', code, message);
      },
      debug(code, ...message) {
        console.debug('NextAuth debug:', code, message);
      },
    },
    callbacks: {
      async jwt({ token, account }) {
        if (account) {
          token.idToken = account.id_token;
          token.accessToken = account.access_token;
          token.refreshToken = account.refresh_token;
          token.expiresAt = account.expires_at;
          token.issuer = cookies().get('kc_issuer')?.value ?? process.env.KEYCLOAK_ISSUER;
          return token;
        }
  
        if (Date.now() < token.expiresAt! * 1000 - 60 * 1000) {
          return token;
        } else {
          try {
            const response = await requestRefreshOfAccessToken(token);
            const tokens: TokenSet = await response.json();
  
            if (!response.ok) throw tokens;
  
            const updatedToken: JWT = {
              ...token,
              idToken: tokens.id_token,
              accessToken: tokens.access_token,
              expiresAt: Math.floor(Date.now() / 1000 + (tokens.expires_in as number)),
              refreshToken: tokens.refresh_token ?? token.refreshToken,
            };
            console.log('Refreshed JWT token:', updatedToken);
            return updatedToken;
          } catch (error) {
            console.error('Error refreshing access token', error);
            return { ...token, error: 'RefreshAccessTokenError' };
          }
        }
      },
      async session({ session, token }) {
        console.log('Using Keycloak issuer:', token);
        console.log("session call back");
        session.accessToken = token.accessToken;
        session.error = token.error;
        session.expiresAt = token.expiresAt;
        return session;
      },
    },
  };

  return authOptions;
}

export default async function handler(req: NextApiRequest, res: NextApiResponse) {

  try {
    const issuerCookie = req.cookies['kc_issuer']; // Access cookies from the request
    
    const issuer = cookies().get('kc_issuer')?.value ?? process.env.KEYCLOAK_ISSUER
  
    const authOptions: AuthOptions = getAuthOptions(issuer)
  
    return await NextAuth(req, res, authOptions);
  } catch (error) {
    console.error('Error in NextAuth handler:', error);
    res.status(500).json({ error: 'Authentication failed' });
  }
}