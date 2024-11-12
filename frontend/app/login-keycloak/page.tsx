import { getServerSession } from 'next-auth';
import { getAuthOptions } from '@/lib/auth';
import ClientHome from '../../components/ClientHome';
import { cookies } from 'next/headers';

export default async function Home() {
  const issuerCookie = cookies().get('kc_issuer');
  const issuer = issuerCookie?.value || process.env.KEYCLOAK_ISSUER;
  // Generate authOptions dynamically
  const authOptions = getAuthOptions(issuer);
  const session = await getServerSession(authOptions);

  const accessToken = session?.accessToken;

  return <ClientHome session={session} accessToken={accessToken} />;
}
