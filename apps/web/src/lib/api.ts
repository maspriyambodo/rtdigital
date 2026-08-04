export interface ApiErrorDetail {
  field?: string;
  issue: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: ApiErrorDetail[];
  request_id?: string;
}

export class ApiException extends Error {
  readonly code: string;
  readonly details: ApiErrorDetail[];
  readonly requestId?: string;
  readonly status: number;

  constructor(status: number, error: ApiError) {
    super(error.message);
    this.name = "ApiException";
    this.status = status;
    this.code = error.code;
    this.details = error.details ?? [];
    this.requestId = error.request_id;
  }
}

const apiBase = (process.env.NEXT_PUBLIC_API_URL || "/api/v1").replace(/\/+$/, "");

export async function apiFetch<T>(path: string, options: RequestInit = {}, unwrapData = true): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body && typeof options.body === "string" && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${apiBase}/${path.replace(/^\/+/, "")}`, {
    ...options,
    headers,
    credentials: "include",
    cache: "no-store",
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new ApiException(
      response.status,
      payload.error ?? {
        code: "HTTP_ERROR",
        message: response.statusText || "Permintaan gagal diproses.",
        request_id: response.headers.get("X-Request-ID") || undefined,
      },
    );
  }

  return (unwrapData ? payload.data ?? payload : payload) as T;
}
