// components/ClientHome.tsx
"use client";

import { useEffect, useState } from "react";
import Logout from "@/components/Logout";
import Login from "@/components/Login";
import { userEmailLoginWithAccessToken } from "@/lib/backend";

export default function ClientHome({ session, accessToken }) {
    const [backendAccessToken, setBackendAccessToken] = useState<string | null>(null);

    useEffect(() => {
        if (accessToken && !backendAccessToken) {
            userEmailLoginWithAccessToken(accessToken, setBackendAccessToken)
                .then(() => {
                    console.log("User logged in with email and token successfully.");
                })
                .catch((error) => {
                    console.error("Failed to log in with email using the access token:", error);
                });
        }
    }, [accessToken, backendAccessToken]);

    return (
        <div className="flex flex-col space-y-3 justify-center items-center h-screen">
            <Logout />
            {!session && (
                <div>
                    <Login />
                </div>
            )}
        </div>
    );
}
