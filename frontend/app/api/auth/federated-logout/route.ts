import { JWT, getToken } from 'next-auth/jwt';
import { NextRequest, NextResponse } from 'next/server';

function logoutParams(req: NextRequest, token: JWT): Record<string, string> {
  const protocol = req.headers.get('x-forwarded-proto') || 'http';
  const host = req.headers.get('x-forwarded-host') || req.headers.get('host') || 'localhost';
  const baseUrl = `${protocol}://${host}`;

  return {
    id_token_hint: token.idToken as string,
    post_logout_redirect_uri: baseUrl,
  };
}

function handleEmptyToken() {
  const response = { error: 'No session present' };
  const responseHeaders = { status: 400 };
  return NextResponse.json(response, responseHeaders);
}

async function sendEndSessionEndpointToURL(req: NextRequest, token: JWT) {
  const endSessionEndPoint = new URL(
    `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/logout`
  );

  const params: Record<string, string> = logoutParams(req, token);
  const endSessionParams = new URLSearchParams(params);

  const response = await fetch(endSessionEndPoint.toString(), {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: endSessionParams.toString(),
    method: 'POST',
    cache: 'no-store',
    redirect: 'manual', 
  });

  return response;
}

function clearAuthCookies(response: NextResponse) {
  const cookies = [
    '__Secure-next-auth.session-token',
    '__Secure-next-auth.session-token.0',
    '__Secure-next-auth.session-token.1',
    '__Host-next-auth.csrf-token',
    '__Secure-next-auth.callback-url',
  ];

  cookies.forEach((cookieName) => {
    response.cookies.set(cookieName, '', {
      path: '/',
      expires: new Date(0),
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
    });
  });
}

export async function GET(req: NextRequest) {
  try {
    const token = await getToken({ req, secret: process.env.NEXTAUTH_SECRET, secureCookie: process.env.NODE_ENV === 'production' });

    if (!token) {
      return handleEmptyToken();
    }

    const response = await sendEndSessionEndpointToURL(req, token);

    console.log("Response: ", response);
    if (response.status === 302) {
      const redirectUrl = response.headers.get('Location');

      console.log("Redirect URL: ", redirectUrl);
      if (redirectUrl) {
        const response = NextResponse.json({ redirectUrl }, { status: 200 });

        clearAuthCookies(response);

        return response;
      } else {
        return NextResponse.json(
          { error: 'No redirect location provided by Keycloak.' },
          { status: 500 }
        );
      }
    }

    return NextResponse.json(
      { error: 'Unexpected response from Keycloak logout endpoint.' },
      { status: response.status }
    );
  } catch (error) {
    console.error('Logout Error:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}

