import Image from "next/image";
import { PengurusNavigation } from "@/components/layout/PengurusNavigation";

export default function PengurusLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        background: "var(--color-bg)",
      }}
    >
      <div className="pengurus-desktop-sidebar" style={{ display: "none" }}>
        <PengurusNavigation />
      </div>

      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          minWidth: 0,
        }}
      >
        <header
          className="pengurus-mobile-header"
          style={{
            position: "sticky",
            top: 0,
            zIndex: 20,
            padding: "var(--space-3) var(--space-4)",
            borderBottom: "1px solid var(--color-border)",
            background: "var(--color-surface)",
            boxShadow: "var(--shadow-sm)",
          }}
        >
          <div
            style={{
              position: "relative",
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              gap: "var(--space-3)",
            }}
          >
            <div
              style={{
                display: "flex",
                minWidth: 0,
                alignItems: "center",
                gap: "var(--space-2)",
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
              <span
                style={{
                  overflow: "hidden",
                  fontSize: "1rem",
                  fontWeight: 700,
                  letterSpacing: "-0.01em",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                RT Digital
              </span>
            </div>
            <PengurusNavigation mobile />
          </div>
        </header>

        <main
          style={{
            flex: 1,
            width: "100%",
            maxWidth: "1280px",
            margin: "0 auto",
            padding: "var(--space-6) var(--space-4)",
            paddingBottom: "calc(var(--space-8) + env(safe-area-inset-bottom))",
          }}
        >
          {children}
        </main>
      </div>

      <style
        dangerouslySetInnerHTML={{
          __html: `
            @media (min-width: 1024px) {
              .pengurus-desktop-sidebar { display: block !important; }
              .pengurus-mobile-header { display: none !important; }
            }
          `,
        }}
      />
    </div>
  );
}