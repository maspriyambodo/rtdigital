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
          <div style={{ position: "relative" }}>
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