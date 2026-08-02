import Image from "next/image";
import { WargaNavigation } from "@/components/layout/WargaNavigation";

export default function WargaLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        paddingBottom: "calc(60px + env(safe-area-inset-bottom))",
      }}
    >
      <header
        style={{
          position: "sticky",
          top: 0,
          zIndex: 20,
          display: "flex",
          alignItems: "center",
          gap: "var(--space-2)",
          padding: "var(--space-3) var(--space-4)",
          borderBottom: "1px solid var(--color-border)",
          background: "var(--color-surface)",
        }}
      >
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
      </header>
      <main
        style={{
          flex: 1,
          width: "100%",
          maxWidth: "1280px",
          margin: "0 auto",
          padding: "var(--space-4)",
        }}
      >
        {children}
      </main>
      <WargaNavigation />
    </div>
  );
}