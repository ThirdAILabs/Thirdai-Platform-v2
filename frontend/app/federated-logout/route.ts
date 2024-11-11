// app/federated-logout/route.ts
import { JWT } from 'next-auth/jwt';
import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';
import { Session } from 'next-auth';

function getBaseUrl(request: NextRequest): string {
  const protocol = request.headers.get('x-forwarded-proto') || 'https';
  const host = request.headers.get('host') || 'localhost';
  return `${protocol}://${host}`;
}

function logoutParams(request: NextRequest, token: JWT | undefined): Record<string, string> {
  const baseUrl = getBaseUrl(request);
  return {
    id_token_hint: token?.idToken as string,
    post_logout_redirect_uri: baseUrl,
  };
}

export async function GET(request: NextRequest) {
  try {
    const session: Session | null = await auth(request as any);
    const token = session?.token;
    if (!token) {
      return NextResponse.redirect(getBaseUrl(request));
    }

    const issuer = session?.token.issuer || process.env.KEYCLOAK_ISSUER!;
    const endSessionEndpoint = new URL(`${issuer}/protocol/openid-connect/logout`);
    const params = logoutParams(request, session.token);
    const endSessionParams = new URLSearchParams(params);
    const redirectUrl = `${endSessionEndpoint.href}?${endSessionParams}`;

    return NextResponse.redirect(redirectUrl);
  } catch (error) {
    console.error('Error during federated logout:', error);
    return NextResponse.redirect(getBaseUrl(request));
  }
}
