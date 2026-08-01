"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";

const items = [
  { href: "/pengurus", label: "Dashboard" },
  { href: "/pengurus/keluarga", label: "Data Warga" },
  { href: "/pengurus/tagihan", label: "Iuran" },
  { href: "/pengurus/kas", label: "Buku Kas" },
  { href: "/pengurus/surat", label: "Surat" },
  { href: "/pengurus/aduan", label: "Aduan" },
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
        gap: "var(--space-2)",
      }}
    >
      {items.map((item) => {
        const isActive =
          pathname === item.href || pathname.startsWith(`${item.href}/`);

        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={isActive ? "page" : undefined}
            onClick={() => setIsOpen(false)}
            style={{
              display: "flex",
              alignItems: "center",
              minHeight: 44,
              padding: "var(--space-2) var(--space-3)",
              borderRadius: "var(--radius-md)",
              color: isActive
                ? "var(--color-primary-700)"
                : "var(--color-text)",
              background: isActive
                ? "var(--color-primary-50)"
                : "transparent",
              fontWeight: isActive ? 600 : 400,
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
          width: 250,
          flexShrink: 0,
          borderRight: "1px solid var(--color-border)",
          background: "var(--color-surface-muted)",
          padding: "var(--space-4)",
        }}
      >
        <div
          style={{
            marginBottom: "var(--space-4)",
            paddingBottom: "var(--space-4)",
            borderBottom: "1px solid var(--color-border)",
            fontWeight: 700,
          }}
        >
          RT Digital
        </div>
        {nav}
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
          minHeight: 44,
          padding: "var(--space-2) var(--space-3)",
          border: "1px solid var(--color-border-strong)",
          borderRadius: "var(--radius-md)",
          background: "var(--color-surface)",
          color: "var(--color-text)",
          fontWeight: 600,
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
            padding: "var(--space-4)",
            border: "1px solid var(--color-border)",
            borderRadius: "var(--radius-md)",
            background: "var(--color-surface)",
            boxShadow: "var(--shadow-md)",
          }}
        >
          {nav}
        </div>
      ) : null}
    </>
  );
}