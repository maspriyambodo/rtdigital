"use client";

import { useRef, useState } from "react";

import { ApiException, apiFetch } from "../../lib/api";
import { Button } from "./Button";

const acceptedMIMETypes = ["image/jpeg", "image/png", "application/pdf"] as const;
const defaultAccept = acceptedMIMETypes.join(",");
const defaultMaxSizeBytes = 10 * 1024 * 1024;

type UploadStatus = "idle" | "uploading" | "confirming" | "success" | "error";

interface PresignUploadResponse {
  file_id: string;
  upload_url: string;
  upload_headers: Record<string, string>;
}

export interface FileUploaderProps {
  entityType?: "payment" | "cash_transaction";
  entityId: string;
  onChange: (fileId: string | undefined) => void;
  disabled?: boolean;
  maxSizeBytes?: number;
}

function uploadToStorage(
  url: string,
  headers: Record<string, string>,
  file: File,
  onProgress: (progress: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();

    request.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    });
    request.addEventListener("load", () => {
      if (request.status >= 200 && request.status < 300) {
        resolve();
        return;
      }
      reject(new Error("Gagal mengunggah file ke penyimpanan."));
    });
    request.addEventListener("error", () => reject(new Error("Koneksi upload gagal.")));
    request.addEventListener("abort", () => reject(new Error("Upload dibatalkan.")));

    request.open("PUT", url);
    for (const [key, value] of Object.entries(headers)) {
      request.setRequestHeader(key, value);
    }
    request.send(file);
  });
}

export function FileUploader({
  entityType = "payment",
  entityId,
  onChange,
  disabled = false,
  maxSizeBytes = defaultMaxSizeBytes,
}: FileUploaderProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File>();
  const [fileID, setFileID] = useState<string>();
  const [status, setStatus] = useState<UploadStatus>("idle");
  const [progress, setProgress] = useState(0);
  const [message, setMessage] = useState<string>();

  const resetInput = () => {
    if (inputRef.current) {
      inputRef.current.value = "";
    }
  };

  const remove = () => {
    setFile(undefined);
    setFileID(undefined);
    setStatus("idle");
    setProgress(0);
    setMessage(undefined);
    resetInput();
    onChange(undefined);
  };

  const upload = async (selectedFile: File) => {
    if (!entityId) {
      setStatus("error");
      setMessage(entityType === "cash_transaction" ? "Transaksi belum siap." : "Tagihan belum dipilih.");
      return;
    }
    if (!acceptedMIMETypes.includes(selectedFile.type as (typeof acceptedMIMETypes)[number])) {
      setStatus("error");
      setMessage("Format bukti harus JPG, PNG, atau PDF.");
      return;
    }
    if (selectedFile.size <= 0 || selectedFile.size > maxSizeBytes) {
      setStatus("error");
      setMessage(`Ukuran bukti maksimal ${Math.floor(maxSizeBytes / 1024 / 1024)} MB.`);
      return;
    }

    try {
      setFile(selectedFile);
      setFileID(undefined);
      setStatus("uploading");
      setProgress(0);
      setMessage(undefined);
      onChange(undefined);

      const presigned = await apiFetch<PresignUploadResponse>("files/presign-upload", {
        method: "POST",
        body: JSON.stringify({
          entity_type: entityType,
          entity_id: entityId,
          purpose: entityType === "cash_transaction" ? "proof" : "payment_proof",
          original_name: selectedFile.name,
          mime_type: selectedFile.type,
          size_bytes: selectedFile.size,
        }),
      });

      await uploadToStorage(
        presigned.upload_url,
        presigned.upload_headers,
        selectedFile,
        setProgress,
      );

      setStatus("confirming");
      await apiFetch("files/confirm-upload", {
        method: "POST",
        body: JSON.stringify({ file_id: presigned.file_id }),
      });

      setFileID(presigned.file_id);
      setStatus("success");
      setProgress(100);
      onChange(presigned.file_id);
    } catch (error) {
      const fallback = entityType === "cash_transaction" ? "Bukti transaksi gagal diunggah." : "Bukti pembayaran gagal diunggah.";
      setStatus("error");
      setMessage(error instanceof ApiException || error instanceof Error ? error.message : fallback);
      setProgress(0);
      onChange(undefined);
    } finally {
      resetInput();
    }
  };

  const selectFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = event.target.files?.[0];
    if (selectedFile) {
      void upload(selectedFile);
    }
  };

  const busy = status === "uploading" || status === "confirming";
  const maxSizeMB = Math.floor(maxSizeBytes / 1024 / 1024);

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "var(--space-2)",
      }}
    >
      <input
        ref={inputRef}
        type="file"
        accept={defaultAccept}
        capture="environment"
        disabled={disabled || busy}
        onChange={selectFile}
        style={{ display: "none" }}
      />

      {status === "success" && fileID ? (
        <>
          <p style={{ margin: 0, color: "var(--color-success)", fontSize: "0.875rem" }}>
            Bukti terunggah: {file?.name}
          </p>
          <Button type="button" variant="outline" disabled={disabled} onClick={remove}>
            Hapus bukti
          </Button>
        </>
      ) : (
        <Button
          type="button"
          variant="secondary"
          disabled={disabled || busy}
          loading={busy}
          onClick={() => inputRef.current?.click()}
        >
          {status === "uploading"
            ? `Mengunggah ${progress}%`
            : status === "confirming"
              ? "Memverifikasi bukti…"
              : status === "error"
                ? "Pilih ulang bukti"
                : "Ambil foto atau pilih bukti"}
        </Button>
      )}

      {busy && (
        <div
          aria-label={`Progress upload ${progress}%`}
          aria-valuemax={100}
          aria-valuemin={0}
          aria-valuenow={progress}
          role="progressbar"
          style={{
            background: "var(--color-surface-muted)",
            borderRadius: "var(--radius-sm)",
            height: 8,
            overflow: "hidden",
          }}
        >
          <div
            style={{
              background: "var(--color-primary-600)",
              height: "100%",
              transition: "width 150ms ease",
              width: `${progress}%`,
            }}
          />
        </div>
      )}

      {status === "error" && (
        <div role="alert" style={{ color: "var(--color-danger)", fontSize: "0.875rem" }}>
          <p style={{ margin: 0 }}>{message}</p>
          {file && (
            <Button
              type="button"
              variant="ghost"
              disabled={disabled}
              onClick={() => void upload(file)}
              style={{ marginTop: "var(--space-1)", padding: "var(--space-1) 0" }}
            >
              Coba lagi
            </Button>
          )}
        </div>
      )}

      <p style={{ color: "var(--color-text-muted)", fontSize: "0.875rem", margin: 0 }}>
        JPG, PNG, atau PDF. Maksimal {maxSizeMB} MB. Kamera tersedia pada perangkat yang mendukung.
      </p>
    </div>
  );
}