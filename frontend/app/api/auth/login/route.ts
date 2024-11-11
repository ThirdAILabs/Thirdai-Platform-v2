// app/api/auth/login/route.ts
import { signIn } from '@/lib/auth';
import { NextRequest, NextResponse } from 'next/server';


export async function GET(req: NextRequest) {
  // Extract host and protocol from the request
  const host = req.headers.get('host');
  const protocol = req.headers.get('x-forwarded-proto');
  const origin = `${protocol}://${host}`;

  // Construct the sign-in URL
  await signIn('keycloak');
}
