import { apiFetch } from "./api";

export type LetterTypeStatus = "active" | "inactive";
export type LetterRequestStatus =
  | "draft"
  | "submitted"
  | "under_review"
  | "needs_revision"
  | "awaiting_approval"
  | "approved"
  | "issued"
  | "rejected"
  | "cancelled";

export interface RequirementItem {
  id?: string;
  name?: string;
  label?: string;
  required: boolean;
  description?: string;
}

export interface FormFieldItem {
  name: string;
  label: string;
  type: "text" | "textarea" | "select" | "date" | "number";
  required: boolean;
  placeholder?: string;
  options?: string[];
}

export interface FormSchemaDefinition {
  fields?: FormFieldItem[];
}

export interface LetterTypeItem {
  id: string;
  name: string;
  requirements: RequirementItem[];
  form_schema: FormSchemaDefinition;
  template: string;
  number_pattern: string;
  status: LetterTypeStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateLetterTypeRequest {
  name: string;
  requirements: RequirementItem[];
  form_schema: FormSchemaDefinition;
  template: string;
  number_pattern: string;
  status: LetterTypeStatus;
}

export interface AttachmentInfo {
  attachment_id: string;
  file_id: string;
  original_name: string;
  mime_type: string;
  size_bytes: number;
  purpose: string;
}

export interface LetterRequestItem {
  id: string;
  requester_user_id: string;
  requester_name: string;
  resident_id: string;
  resident_name: string;
  letter_type_id: string;
  letter_type_name: string;
  request_number: string;
  letter_number?: string;
  form_data: Record<string, unknown>;
  status: LetterRequestStatus;
  resident_note?: string;
  internal_note?: string;
  submitted_at?: string;
  processed_by?: string;
  approved_by?: string;
  approved_at?: string;
  issued_file_id?: string;
  issued_at?: string;
  created_at: string;
  updated_at: string;
  attachments: AttachmentInfo[];
}

export interface SubmitLetterRequest {
  letter_type_id: string;
  resident_id: string;
  form_data: Record<string, unknown>;
  attachment_file_ids: string[];
  resident_note?: string;
}

export interface ReviewLetterRequest {
  resident_note?: string;
  internal_note?: string;
}

export interface LetterRequestFilter {
  status?: LetterRequestStatus;
  letter_type_id?: string;
  search?: string;
}

export interface DownloadLetterResponse {
  file_id: string;
  download_url: string;
  expires_at: string;
}

const options = (token: string, init: RequestInit = {}): RequestInit => ({
  ...init,
  headers: { ...init.headers, Authorization: `Bearer ${token}` },
});

const queryPath = <T extends object>(path: string, values: T) => {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== "") query.set(key, String(value));
  }
  return query.size ? `${path}?${query}` : path;
};

export const listLetterTypes = (token: string, includeInactive = false) =>
  apiFetch<LetterTypeItem[]>(includeInactive ? "letter-types?include_inactive=true" : "letter-types", options(token));

export const getLetterType = (token: string, id: string) =>
  apiFetch<LetterTypeItem>(`letter-types/${id}`, options(token));

export const createLetterType = (token: string, data: CreateLetterTypeRequest) =>
  apiFetch<LetterTypeItem>("letter-types", options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateLetterType = (token: string, id: string, data: CreateLetterTypeRequest) =>
  apiFetch<LetterTypeItem>(`letter-types/${id}`, options(token, { method: "PATCH", body: JSON.stringify(data) }));

export const deactivateLetterType = (token: string, id: string) =>
  apiFetch<LetterTypeItem>(`letter-types/${id}/deactivate`, options(token, { method: "POST" }));

export const listLetterRequests = (token: string, filter: LetterRequestFilter = {}) =>
  apiFetch<LetterRequestItem[]>(queryPath("letter-requests", filter), options(token));

export const submitLetterRequest = (token: string, data: SubmitLetterRequest) =>
  apiFetch<LetterRequestItem>("letter-requests", options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateLetterRequest = (token: string, id: string, data: SubmitLetterRequest) =>
  apiFetch<LetterRequestItem>(`letter-requests/${id}`, options(token, { method: "PATCH", body: JSON.stringify(data) }));

export const processLetterRequest = (token: string, id: string, data: ReviewLetterRequest = {}) =>
  apiFetch<LetterRequestItem>(`letter-requests/${id}/process`, options(token, { method: "POST", body: JSON.stringify(data) }));

export const requestRevision = (token: string, id: string, data: ReviewLetterRequest) =>
  apiFetch<LetterRequestItem>(`letter-requests/${id}/request-revision`, options(token, { method: "POST", body: JSON.stringify(data) }));

export const approveLetterRequest = (token: string, id: string, data: ReviewLetterRequest = {}) =>
  apiFetch<LetterRequestItem>(`letter-requests/${id}/approve`, options(token, { method: "POST", body: JSON.stringify(data) }));

export const rejectLetterRequest = (token: string, id: string, data: ReviewLetterRequest) =>
  apiFetch<LetterRequestItem>(`letter-requests/${id}/reject`, options(token, { method: "POST", body: JSON.stringify(data) }));

export const issueLetter = (token: string, id: string) =>
  apiFetch<LetterRequestItem>(`letter-requests/${id}/issue`, options(token, { method: "POST" }));

export const downloadLetter = (token: string, id: string) =>
  apiFetch<DownloadLetterResponse>(`letter-requests/${id}/download`, options(token));