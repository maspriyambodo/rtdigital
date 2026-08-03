import type { Metadata, Viewport } from "next";
import { AuthProvider } from "@/lib/auth-context";
import { AuthGuard } from "@/components/layout/AuthGuard";
import { DynamicTitle } from "@/components/layout/DynamicTitle";
import { OfflineNotice } from "@/components/layout/OfflineNotice";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "RT Digital",
    template: "%s | RT Digital",
  },
  description: "Aplikasi layanan warga dan pengurus RT.",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "default",
    title: "RT Digital",
  },
  formatDetection: {
    telephone: false,
  },
};

export const viewport: Viewport = {
  themeColor: "#ffffff",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body>
        <DynamicTitle />
        <AuthProvider>
          <AuthGuard>{children}</AuthGuard>
        </AuthProvider>
        <OfflineNotice />
      </body>
    </html>
  );
}
