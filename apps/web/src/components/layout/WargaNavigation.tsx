"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { fetchNotifications } from "@/lib/notifications";

type IconName =
  | "home"
  | "receipt"
  | "mail"
  | "message"
  | "more"
  | "users"
  | "bell"
  | "user";

type NavigationItem = {
  href: string;
  label: string;
  icon: IconName;
};

const primaryItems: readonly NavigationItem[] = [
  { href: "/warga", label: "Beranda", icon: "home" },
  { href: "/warga/tagihan", label: "Tagihan", icon: "receipt" },
  { href: "/warga/surat", label: "Surat", icon: "mail" },
  { href: "/warga/aduan", label: "Aduan", icon: "message" },
] as const;

const moreItems: readonly NavigationItem[] = [
  { href: "/warga/keluarga", label: "Keluarga", icon: "users" },
  { href: "/warga/pengumuman", label: "Informasi", icon: "bell" },
  { href: "/warga/notifikasi", label: "Notifikasi", icon: "bell" },
  { href: "/warga/profil", label: "Profil", icon: "user" },
] as const;

function Icon({ name }: { name: IconName }) {
  const paths: Record<IconName, React.ReactNode> = {
    home: (
      <>
        <path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1V10Z" />
      </>
    ),
    receipt: (
      <>
        <path d="M6 3h12v18l-3-2-3 2-3-2-3 2V3Z" />
        <path d="M9 8h6M9 12h6M9 16h3" />
      </>
    ),
    mail: (
      <>
        <rect x="3" y="5" width="18" height="14" rx="2" />
        <path d="m3 7 9 6 9-6" />
      </>
    ),
    message: (
      <>
        <path d="M21 11.5a8.4 8.4 0 0 1-9 8.5 9.7 9.7 0 0 1-4-.9L3 21l1.7-4.1A8.1 8.1 0 0 1 3 11.5C3 6.8 7 3 12 3s9 3.8 9 8.5Z" />
      </>
    ),
    more: <path d="M5 12h.01M12 12h.01M19 12h.01" />,
    users: (
      <>
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M22 21v-2a4 4 0 0 0-3-3.9M16 3.1a4 4 0 0 1 0 7.8" />
      </>
    ),
    bell: (
      <>
        <path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4" />
      </>
    ),
    user: (
      <>
        <circle cx="12" cy="7" r="4" />
        <path d="M5 21a7 7 0 0 1 14 0" />
      </>
    ),
  };

  return (
    <svg
      aria-hidden="true"
      width="22"
      height="22"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {paths[name]}
    </svg>
  );
}

function isItemActive(pathname: string, href: string) {
  return pathname === href || (href !== "/warga" && pathname.startsWith(`${href}/`));
}

export function WargaNavigation() {
  const pathname = usePathname();
  const { getAccessToken, isInitialized } = useAuth();
  const [unreadCount, setUnreadCount] = useState(0);
  const [isMoreOpen, setIsMoreOpen] = useState(false);
  const isMoreActive = moreItems.some((item) => isItemActive(pathname, item.href));

  useEffect(() => {
    if (!isInitialized) return;

    let active = true;
    void (async () => {
      const token = await getAccessToken();
      if (!token) return;
      try {
        const notifications = await fetchNotifications(token, true);
        if (active) setUnreadCount(notifications.length);
      } catch {
        // Badge must not block navigation when the notification API is unavailable.
      }
    })();

    return () => {
      active = false;
    };
  }, [getAccessToken, isInitialized, pathname]);

  return (
    <>
      {isMoreOpen ? (
        <div
          id="warga-menu-lainnya"
          aria-label="Menu lainnya"
          style={{
            position: "fixed",
            right: "var(--space-4)",
            bottom: "calc(72px + env(safe-area-inset-bottom))",
            zIndex: 41,
            display: "grid",
            minWidth: 184,
            overflow: "hidden",
            border: "1px solid var(--color-border)",
            borderRadius: "var(--radius-lg)",
            background: "var(--color-surface)",
            boxShadow: "var(--shadow-lg)",
          }}
        >
          {moreItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setIsMoreOpen(false)}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "var(--space-3)",
                minHeight: 48,
                padding: "0 var(--space-4)",
                color: isItemActive(pathname, item.href)
                  ? "var(--color-primary-700)"
                  : "var(--color-text)",
                fontWeight: isItemActive(pathname, item.href) ? 700 : 500,
                textDecoration: "none",
              }}
            >
              <Icon name={item.icon} />
              <span>{item.label}</span>
              {item.href === "/warga/notifikasi" && unreadCount > 0 ? (
                <span
                  aria-label={`${unreadCount} notifikasi belum dibaca`}
                  style={{
                    marginLeft: "auto",
                    minWidth: 20,
                    padding: "1px 6px",
                    borderRadius: "var(--radius-full)",
                    background: "var(--color-danger-600)",
                    color: "#ffffff",
                    fontSize: "0.6875rem",
                    fontWeight: 700,
                    textAlign: "center",
                  }}
                >
                  {unreadCount > 9 ? "9+" : unreadCount}
                </span>
              ) : null}
            </Link>
          ))}
        </div>
      ) : null}

      <nav
        aria-label="Navigasi warga"
        style={{
          position: "fixed",
          right: 0,
          bottom: 0,
          left: 0,
          zIndex: 40,
          display: "grid",
          gridTemplateColumns: "repeat(5, 1fr)",
          borderTop: "1px solid var(--color-border)",
          background: "var(--color-surface)",
          paddingBottom: "env(safe-area-inset-bottom)",
          boxShadow: "0 -4px 12px rgb(15 23 42 / 0.05)",
        }}
      >
        {primaryItems.map((item) => {
          const isActive = isItemActive(pathname, item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              aria-label={item.label}
              aria-current={isActive ? "page" : undefined}
              title={item.label}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                minHeight: 56,
                borderTop: isActive
                  ? "2px solid var(--color-primary-600)"
                  : "2px solid transparent",
                color: isActive
                  ? "var(--color-primary-700)"
                  : "var(--color-text-secondary)",
                transition:
                  "border-color var(--transition-fast), color var(--transition-fast)",
              }}
            >
              <Icon name={item.icon} />
            </Link>
          );
        })}

        <button
          type="button"
          aria-label="Menu lainnya"
          aria-expanded={isMoreOpen}
          aria-controls="warga-menu-lainnya"
          title="Menu lainnya"
          onClick={() => setIsMoreOpen((open) => !open)}
          style={{
            position: "relative",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            minHeight: 56,
            border: 0,
            borderTop: isMoreActive
              ? "2px solid var(--color-primary-600)"
              : "2px solid transparent",
            background: "transparent",
            color: isMoreActive
              ? "var(--color-primary-700)"
              : "var(--color-text-secondary)",
            cursor: "pointer",
          }}
        >
          <Icon name="more" />
          {unreadCount > 0 ? (
            <span
              aria-label={`${unreadCount} notifikasi belum dibaca`}
              style={{
                position: "absolute",
                top: 7,
                right: "calc(50% - 17px)",
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
        </button>
      </nav>
    </>
  );
}