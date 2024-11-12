'use client';
import { useEffect } from 'react';
import Cookies from 'js-cookie';
export default function KeycloakIssuerSetter() {
    useEffect(() => {
        const currentDomain = typeof window !== 'undefined' ? window.location.origin : null;
        const dynamicIssuer = currentDomain
            ? `${currentDomain}/keycloak/realms/ThirdAI-Platform`
            : process.env.KEYCLOAK_ISSUER; // fallback from env var if `window` is not available

        Cookies.set('kc_issuer', dynamicIssuer, {
            secure: true,
            sameSite: 'strict',
            path: '/',
        });
    }, []);
    return null;
}