"use client";

import { useEffect, useState } from "react";

export function OfflineNotice() {
  const [isOffline, setIsOffline] = useState(false);

  useEffect(() => {
    const updateStatus = () => setIsOffline(!navigator.onLine);

    updateStatus();
    window.addEventListener("online", updateStatus);
    window.addEventListener("offline", updateStatus);

    return () => {
      window.removeEventListener("online", updateStatus);
      window.removeEventListener("offline", updateStatus);
    };
  }, []);

  if (!isOffline) {
    return null;
  }

  return (
    <div
      role="alert"
      style={{
        position: "fixed",
        right: "var(--space-4)",
        bottom: "calc(var(--space-4) + env(safe-area-inset-bottom))",
        left: "var(--space-4)",
        zIndex: 50,
        padding: "var(--space-3) var(--space-4)",
        borderRadius: "var(--radius-md)",
        background: "var(--color-warning-bg)",
        color: "var(--color-warning)",
        boxShadow: "var(--shadow-md)",
        fontWeight: 500,
        textAlign: "center",
      }}
    >
      Anda sedang luring. Beberapa fitur mungkin tidak tersedia.
    </div>
  );
}