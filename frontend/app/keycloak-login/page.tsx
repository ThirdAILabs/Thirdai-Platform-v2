import { getServerSession } from 'next-auth'
import { authOptions } from '../../app/api/auth/[...nextauth]/route'
import Login from '../../components/Login'
import Logout from '../../components/Logout'


export default async function Home() {
    console.log("KEYCLOAK_ISSUER:", process.env.KEYCLOAK_ISSUER);
    console.log("NEXTAUTH_SECRET:", process.env.NEXTAUTH_SECRET);

    console.log("Auth:", authOptions)
    const session = await getServerSession(authOptions)
    console.log(session)
    console.log("Access Token: ", session?.accessToken)
    if (session?.accessToken) {
        return <div className='flex flex-col space-y-3 justify-center items-center h-screen'>
            <div>Your name is {session.user?.name}</div>
            <div>
                <Logout />
            </div>
        </div>
    }
    console.log("Session is Login");
    return (
        <div className='flex justify-center items-center h-screen'>
            <Login />
        </div>
    )
}
