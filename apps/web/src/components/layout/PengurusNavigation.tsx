"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

const items = [
  { href: "/pengurus", label: "Dashboard" },
  { href: "/pengurus/rumah", label: "Rumah & Unit" },
  { href: "/pengurus/keluarga", label: "Data Keluarga" },
  { href: "/pengurus/warga", label: "Data Warga" },
  { href: "/pengurus/pengguna", label: "Pengguna & Peran" },
  { href: "/pengurus/pengumuman", label: "Pengumuman & Agenda" },
  { href: "/pengurus/tagihan", label: "Iuran" },
  { href: "/pengurus/kas", label: "Buku Kas" },
  { href: "/pengurus/surat", label: "Surat" },
  { href: "/pengurus/aduan", label: "Aduan" },
  { href: "/pengurus/pengaturan", label: "Pengaturan RT" },
  { href: "/pengurus/audit", label: "Audit Log" },
] as const;

type PengurusNavigationProps = {
  mobile?: boolean;
};

export function PengurusNavigation({
  mobile = false,
}: PengurusNavigationProps) {
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);

  const nav = (
    <nav
      aria-label="Navigasi pengurus"
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "var(--space-1)",
      }}
    >
      {items.map((item) => {
        const isActive =
          pathname === item.href ||
          (item.href !== "/pengurus" && pathname.startsWith(`${item.href}/`));

        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={isActive ? "page" : undefined}
            onClick={() => setIsOpen(false)}
            style={{
              display: "flex",
              alignItems: "center",
              minHeight: 40,
              padding: "var(--space-2) var(--space-3)",
              border: isActive
                ? "1px solid var(--color-primary-100)"
                : "1px solid transparent",
              borderRadius: "var(--radius-md)",
              background: isActive
                ? "var(--color-primary-50)"
                : "transparent",
              color: isActive
                ? "var(--color-primary-700)"
                : "var(--color-text-secondary)",
              fontSize: "0.875rem",
              fontWeight: isActive ? 600 : 500,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
              transition:
                "background-color var(--transition-fast), color var(--transition-fast), border-color var(--transition-fast)",
            }}
          >
            {item.label}
          </Link>
        );
      })}
    </nav>
  );

  if (!mobile) {
    return (
      <aside
        aria-label="Panel pengurus"
        style={{
          position: "sticky",
          top: 0,
          display: "flex",
          flexDirection: "column",
          gap: "var(--space-5)",
          width: 260,
          height: "100vh",
          flexShrink: 0,
          padding: "var(--space-5)",
          borderRight: "1px solid var(--color-border)",
          background: "var(--color-surface)",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "var(--space-3)",
            paddingBottom: "var(--space-4)",
            borderBottom: "1px solid var(--color-border)",
          }}
        >
          <Image
            src="/logo.png"
            alt=""
            width={36}
            height={36}
            priority
            style={{
              flexShrink: 0,
              width: "auto",
              height: "2.25rem",
              maxWidth: "100%",
              objectFit: "contain",
            }}
          />
          <div>
            <span
              style={{
                display: "block",
                fontSize: "1rem",
                fontWeight: 700,
                letterSpacing: "-0.01em",
              }}
            >
              RT Digital
            </span>
            <span
              style={{
                display: "block",
                color: "var(--color-text-muted)",
                fontSize: "0.75rem",
              }}
            >
              Portal Pengurus
            </span>
          </div>
        </div>
        <div style={{ flex: 1, overflowY: "auto" }}>{nav}</div>
      </aside>
    );
  }

  return (
    <>
      <button
        type="button"
        aria-expanded={isOpen}
        aria-controls="pengurus-mobile-navigation"
        onClick={() => setIsOpen((open) => !open)}
        style={{
          minHeight: 40,
          padding: "var(--space-2) var(--space-3)",
          border: "1px solid var(--color-border-strong)",
          borderRadius: "var(--radius-md)",
          background: "var(--color-surface)",
          color: "var(--color-text)",
          fontSize: "0.875rem",
          fontWeight: 600,
          boxShadow: "var(--shadow-sm)",
        }}
      >
        Menu pengurus
      </button>

      {isOpen ? (
        <div
          id="pengurus-mobile-navigation"
          style={{
            position: "absolute",
            top: "calc(100% + var(--space-2))",
            right: 0,
            left: 0,
            zIndex: 30,
            maxHeight: "75vh",
            overflowY: "auto",
            padding: "var(--space-3)",
            border: "1px solid var(--color-border)",
            borderRadius: "var(--radius-lg)",
            background: "var(--color-surface)",
            boxShadow: "var(--shadow-lg)",
          }}
        >
          {nav}
        </div>
      ) : null}
    </>
  );
}