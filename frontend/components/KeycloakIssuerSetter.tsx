'use client';
import { useEffect } from 'react';
import Cookies from 'js-cookie';
export default function KeycloakIssuerSetter() {
    useEffect(() => {
        const currentDomain = window.location.origin;
        const dynamicIssuer = `${currentDomain}/keycloak/realms/ThirdAI-Platform`;
        Cookies.set('kc_issuer', dynamicIssuer, {
            secure: true,
            sameSite: 'strict',
            path: '/',
        });
    }, []);
    return null;
}