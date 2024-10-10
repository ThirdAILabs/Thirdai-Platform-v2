// app/keycloak-login/page.tsx
import { getServerSession } from "next-auth";
import { authOptions } from "@/app/api/auth/[...nextauth]/route";
import ClientHome from "../../components/ClientHome"

export default async function Home() {
    const session = await getServerSession(authOptions);

    // Now we pass the session as a prop to client components.
    const accessToken = session?.accessToken;

    return (
        <ClientHome session={session} accessToken={accessToken} />
    );
}
