"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const items = [
  { href: "/warga", label: "Beranda" },
  { href: "/warga/tagihan", label: "Tagihan" },
  { href: "/warga/surat", label: "Surat" },
  { href: "/warga/aduan", label: "Aduan" },
  { href: "/warga/profil", label: "Profil" },
] as const;

export function WargaNavigation() {
  const pathname = usePathname();

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
        gridTemplateColumns: `repeat(${items.length}, 1fr)`,
        borderTop: "1px solid var(--color-border)",
        background: "var(--color-surface)",
        paddingBottom: "env(safe-area-inset-bottom)",
      }}
    >
      {items.map((item) => {
        const isActive = pathname === item.href;

        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={isActive ? "page" : undefined}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              minHeight: 44,
              padding: "var(--space-2) 2px",
              color: isActive
                ? "var(--color-primary-600)"
                : "var(--color-text-secondary)",
              fontSize: "0.75rem",
              fontWeight: isActive ? 600 : 400,
              textAlign: "center",
              whiteSpace: "nowrap",
            }}
          >
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}