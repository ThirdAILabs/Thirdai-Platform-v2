'use client';

import { useContext, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { useRouter } from 'next/navigation'; // Use 'next/navigation' instead of 'next/router' in the app directory
import { userEmailLogin } from '@/lib/backend';
import { UserContext } from '../user_wrapper';
import axios from 'axios';

export default function LoginPage() {
  const [error, setError] = useState<string>('');
  const { setAccessToken } = useContext(UserContext);
  const router = useRouter();

  const handleKeycloakLogin = async () => {
    console.log('Starting Keycloak login process...');

    try {
      // Step 1: Redirect to Keycloak authorization URL
      const authorizationUrl = `http://localhost:8180/realms/new-realm/protocol/openid-connect/auth?client_id=new-client&redirect_uri=http://localhost:80&response_type=code&scope=openid`;
      console.log('Redirecting to Keycloak authorization URL:', authorizationUrl);

      // Redirect the user to Keycloak for authentication
      window.location.href = authorizationUrl;
    } catch (error) {
      console.error('Error during Keycloak login setup:', error);
      setError('An error occurred during login. Please try again.');
    }
  };

  // Step 2: Handle the token exchange after redirect back
  const exchangeCodeForToken = async () => {
    try {
      const code = new URLSearchParams(window.location.search).get('code');
      console.log('Code from URL:', code);

      if (!code) {
        console.error('Authorization code not found in URL.');
        return;
      }

      console.log('Exchanging authorization code for tokens...');
      const response = await axios.post(
        'http://localhost:8180/realms/new-realm/protocol/openid-connect/token',
        new URLSearchParams({
          client_id: 'new-client',
          grant_type: 'authorization_code',
          code: code,
          redirect_uri: 'http://localhost:80',
        }),
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
        }
      );

      const { access_token } = response.data;
      console.log('Successfully received access token:', access_token);

      console.log('Logging in with access token...');
      await userEmailLogin(access_token, setAccessToken);
      console.log('Login successful, redirecting to home page.');

      router.push('/');
    } catch (error) {
      console.error('Error exchanging code for token:', error);
      setError('Login failed. Please try again.');
    }
  };

  // Check for the authorization code directly when the component mounts
  if (typeof window !== 'undefined' && new URLSearchParams(window.location.search).has('code')) {
    exchangeCodeForToken();
  }

  return (
    <div className="min-h-screen flex justify-center items-start md:items-center p-8">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-2xl">Login</CardTitle>
          <CardDescription>Please click the button below to log in using Keycloak.</CardDescription>
        </CardHeader>
        <CardFooter>
          <div className="w-full">
            {error && <p className="text-red-500 text-sm mb-4">{error}</p>}

            <Button onClick={handleKeycloakLogin} className="w-full">
              Login with Keycloak
            </Button>
          </div>
        </CardFooter>
      </Card>
    </div>
  );
}
