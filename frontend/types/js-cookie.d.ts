declare module 'js-cookie' {
    const Cookies: {
        get: (key: string) => string | undefined;
        set: (key: string, value: string, options?: { secure?: boolean; sameSite?: 'strict' | 'lax' | 'none'; path?: string }) => void;
        remove: (key: string) => void;
    };
    export default Cookies;
}
