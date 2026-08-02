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
            padding: "var(--space-4)",
            borderBottom: "1px solid var(--color-border)",
            background: "var(--color-surface)",
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
            <div style={{ display: "flex", minWidth: 0, alignItems: "center", gap: "var(--space-2)" }}>
              <Image
                src="/logo.png"
                alt=""
                width={40}
                height={40}
                priority
                style={{ flexShrink: 0, width: "auto", height: "2.5rem", maxWidth: "100%", objectFit: "contain" }}
              />
              <span style={{ overflow: "hidden", fontWeight: 700, textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
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
            padding: "var(--space-4)",
            paddingBottom: "calc(var(--space-4) + env(safe-area-inset-bottom))",
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