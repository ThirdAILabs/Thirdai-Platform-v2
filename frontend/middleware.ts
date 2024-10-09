export { default } from "next-auth/middleware"

// Don't invoke Middleware on some paths
export const config = {
  matcher: ["/private"],
};
