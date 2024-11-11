// next-auth.d.ts
import NextAuth, { DefaultSession } from 'next-auth';
import { JWT } from 'next-auth/jwt';


declare module 'next-auth/jwt' {
  interface JWT {
    idToken?: string;
    accessToken?: string;
    refreshToken?: string;
    expiresAt?: number;
    error?: string;
    issuer?: string;
  }
}


declare module 'next-auth' {
  interface Session extends DefaultSession {
    token: JWT;
    error?: string;
    expiresAt?: number;
  }
}
