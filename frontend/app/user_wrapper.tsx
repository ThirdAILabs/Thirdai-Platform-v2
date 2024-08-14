"use client"

import { validAccessToken } from "@/lib/backend";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function UserWrapper({children} : {children: React.ReactNode}) {
  const router = useRouter();

  useEffect(() => {
    validAccessToken().then(isValid => {
      if (!isValid) {
        router.push("/login-email");
      }
    });
  }, []);

  return <>
    {children}
  </>
}