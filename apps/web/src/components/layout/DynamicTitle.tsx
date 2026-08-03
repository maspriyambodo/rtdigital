"use client";

import { useEffect } from "react";
import { usePathname } from "next/navigation";

const titles: Record<string, string> = {
  "/login": "Masuk",
  "/activate": "Aktivasi Akun",
  "/forgot-password": "Lupa Kata Sandi",
  "/reset-password": "Reset Kata Sandi",

  "/pengurus": "Dashboard Pengurus",
  "/pengurus/rumah": "Rumah & Unit",
  "/pengurus/keluarga": "Data Keluarga",
  "/pengurus/warga": "Data Warga",
  "/pengurus/pengguna": "Pengguna & Peran",
  "/pengurus/pengguna/invite": "Undang Pengguna",
  "/pengurus/pengumuman": "Pengumuman & Agenda",
  "/pengurus/tagihan": "Iuran & Tagihan",
  "/pengurus/kas": "Buku Kas",
  "/pengurus/surat": "Surat",
  "/pengurus/aduan": "Aduan",
  "/pengurus/pengaturan": "Pengaturan RT",
  "/pengurus/audit": "Audit Log",
  "/pengurus/laporan": "Laporan",

  "/warga": "Beranda",
  "/warga/keluarga": "Keluarga",
  "/warga/pengumuman": "Informasi",
  "/warga/tagihan": "Tagihan",
  "/warga/surat": "Surat",
  "/warga/aduan": "Aduan",
  "/warga/notifikasi": "Notifikasi",
  "/warga/profil": "Profil",
};

function getTitle(pathname: string) {
  if (titles[pathname]) return titles[pathname];

  if (pathname.startsWith("/pengurus/warga/")) return "Detail Data Warga";
  if (pathname.startsWith("/pengurus/pengguna/")) return "Detail Pengguna";

  const parentPath = pathname.split("/").slice(0, 3).join("/");
  return titles[parentPath] ?? "RT Digital";
}

export function DynamicTitle() {
  const pathname = usePathname();

  useEffect(() => {
    const title = getTitle(pathname);
    document.title = title === "RT Digital" ? title : `${title} | RT Digital`;
  }, [pathname]);

  return null;
}