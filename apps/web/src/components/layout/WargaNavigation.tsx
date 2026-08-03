"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { fetchNotifications } from "@/lib/notifications";

const items = [
  { href: "/warga", label: "Beranda" },
  { href: "/warga/keluarga", label: "Keluarga" },
  { href: "/warga/pengumuman", label: "Informasi" },
  { href: "/warga/tagihan", label: "Tagihan" },
  { href: "/warga/surat", label: "Surat" },
  { href: "/warga/aduan", label: "Aduan" },
  { href: "/warga/notifikasi", label: "Notifikasi" },
  { href: "/warga/profil", label: "Profil" },
] as const;

export function WargaNavigation() {
  const pathname = usePathname();
  const { getAccessToken, isInitialized } = useAuth();
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    if (!isInitialized) return;

    let active = true;
    void (async () => {
      const token = await getAccessToken();
      if (!token) return;
      try {
        const items = await fetchNotifications(token, true);
        if (active) setUnreadCount(items.length);
      } catch {
        // Badge must not block navigation when the notification API is unavailable.
      }
    })();

    return () => {
      active = false;
    };
  }, [getAccessToken, isInitialized, pathname]);

  return (
    <nav
      aria-label="Navigasi warga"
      style={{
        position: "fixed",
        right: 0,
        bottom: 0,
        left: 0,
        zIndex: 40,
        display: "grid",
        gridTemplateColumns: "repeat(4, 1fr)",
        gridAutoRows: "1fr",
        borderTop: "1px solid var(--color-border)",
        background: "var(--color-surface)",
        paddingBottom: "env(safe-area-inset-bottom)",
        boxShadow: "0 -4px 12px rgb(15 23 42 / 0.05)",
      }}
    >
      {items.map((item) => {
        const isActive =
          pathname === item.href ||
          (item.href !== "/warga" && pathname.startsWith(`${item.href}/`));

        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={isActive ? "page" : undefined}
            style={{
              position: "relative",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              minHeight: 48,
              padding: "var(--space-2) 2px",
              borderTop: isActive
                ? "2px solid var(--color-primary-600)"
                : "2px solid transparent",
              color: isActive
                ? "var(--color-primary-700)"
                : "var(--color-text-secondary)",
              fontSize: "0.75rem",
              fontWeight: isActive ? 600 : 500,
              textAlign: "center",
              whiteSpace: "nowrap",
              transition:
                "border-color var(--transition-fast), color var(--transition-fast)",
            }}
          >
            {item.label}
            {item.href === "/warga/notifikasi" && unreadCount > 0 ? (
              <span
                aria-label={`${unreadCount} notifikasi belum dibaca`}
                style={{
                  position: "absolute",
                  top: 4,
                  right: 4,
                  minWidth: 16,
                  height: 16,
                  padding: "0 4px",
                  borderRadius: "var(--radius-full)",
                  background: "var(--color-danger-600)",
                  color: "#ffffff",
                  fontSize: "0.625rem",
                  fontWeight: 700,
                  lineHeight: "16px",
                  boxShadow: "0 0 0 2px var(--color-surface)",
                }}
              >
                {unreadCount > 9 ? "9+" : unreadCount}
              </span>
            ) : null}
          </Link>
        );
      })}
    </nav>
  );
}