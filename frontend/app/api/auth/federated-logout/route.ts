import { JWT, getToken } from 'next-auth/jwt';
import { NextRequest, NextResponse } from 'next/server';

function logoutParams(token: JWT): Record<string, string> {
  return {
    id_token_hint: token.idToken as string,
    post_logout_redirect_uri: process.env.NEXTAUTH_URL,
  };
}

function handleEmptyToken() {
  const response = { error: 'No session present' };
  const responseHeaders = { status: 400 };
  return NextResponse.json(response, responseHeaders);
}

async function sendEndSessionEndpointToURL(token: JWT) {
  const endSessionEndPoint = new URL(
    `${process.env.KEYCLOAK_ISSUER}/protocol/openid-connect/logout`
  );

  const params: Record<string, string> = logoutParams(token);
  const endSessionParams = new URLSearchParams(params);

  const response = await fetch(endSessionEndPoint.toString(), {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: endSessionParams.toString(),
    method: 'POST',
    cache: 'no-store',
    redirect: 'manual', // Prevents automatic following of redirects
  });

  return response;
}

function clearAuthCookies(response: NextResponse) {
  const cookies = [
    'next-auth.session-token',
    'next-auth.csrf-token',
    // Add any other NextAuth-related cookies here
  ];

  cookies.forEach((cookieName) => {
    response.cookies.set(cookieName, '', {
      path: '/',
      expires: new Date(0), // Set expiration to a past date to delete the cookie
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
    });
  });
}

export async function GET(req: NextRequest) {
  try {
    // Retrieve the token from the request
    const token = await getToken({ req });
    console.log("Request: ", req);

    // If no token is found, return a 400 error
    if (!token) {
      return handleEmptyToken();
    }

    // Send the logout request to Keycloak
    const response = await sendEndSessionEndpointToURL(token);

    console.log("Response: ", response);
    // Check if Keycloak responded with a 302 redirect
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

    // If Keycloak didn't respond with a 302, handle accordingly
    return NextResponse.json(
      { error: 'Unexpected response from Keycloak logout endpoint.' },
      { status: response.status }
    );
  } catch (error) {
    console.error('Logout Error:', error);
    // Return a 500 Internal Server Error response in case of unexpected errors
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}

