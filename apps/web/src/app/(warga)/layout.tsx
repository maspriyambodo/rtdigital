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