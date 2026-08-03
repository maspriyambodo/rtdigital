"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

const publicPaths = ["/login", "/forgot-password", "/reset-password", "/activate"];

function isPublicPath(pathname: string) {
  return publicPaths.some((path) => pathname === path || pathname.startsWith(`${path}/`));
}

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { accessToken, isInitialized } = useAuth();
  const isPublic = isPublicPath(pathname);

  useEffect(() => {
    if (isInitialized && !accessToken && !isPublic) {
      router.replace("/login");
    }
  }, [accessToken, isInitialized, isPublic, router]);

  if (!isInitialized || (!accessToken && !isPublic)) {
    return null;
  }

  return <>{children}</>;
}