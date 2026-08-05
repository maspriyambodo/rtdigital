import { apiFetch } from "./api";

export type ComplaintPriority = "low" | "normal" | "high";
export type ComplaintStatus =
  | "new"
  | "reviewed"
  | "in_progress"
  | "waiting_information"
  | "resolved"
  | "rejected"
  | "closed";

export interface ComplaintCategory {
  id: string;
  code: string;
  name: string;
  status: "active" | "inactive";
  created_at: string;
  updated_at: string;
}

export interface AttachmentInfo {
  attachment_id: string;
  file_id: string;
  original_name: string;
  mime_type: string;
  size_bytes: number;
  purpose: string;
}

export interface CommentItem {
  id: string;
  complaint_id: string;
  author_user_id: string;
  author_name: string;
  body: string;
  is_internal: boolean;
  created_at: string;
}

export interface ComplaintItem {
  id: string;
  reporter_user_id: string;
  reporter_name: string;
  ticket_number: string;
  complaint_category_id: string;
  category_name: string;
  title: string;
  description: string;
  location_description?: string;
  priority: ComplaintPriority;
  status: ComplaintStatus;
  assigned_to?: string;
  assigned_to_name?: string;
  resolution_note?: string;
  resolved_at?: string;
  closed_at?: string;
  created_at: string;
  updated_at: string;
  attachments: AttachmentInfo[];
  comments: CommentItem[];
}

export interface CreateComplaintRequest {
  complaint_category_id: string;
  title: string;
  description: string;
  location_description?: string;
  priority: ComplaintPriority;
  attachment_file_ids: string[];
}

export type UpdateComplaintRequest = CreateComplaintRequest;

export interface ComplaintFilter {
  status?: ComplaintStatus;
  complaint_category_id?: string;
  assigned_to?: string;
  search?: string;
}

export interface CreateComplaintCategoryRequest {
  code: string;
  name: string;
}

export interface UpdateComplaintCategoryRequest {
  name: string;
  status?: "active" | "inactive";
}

export interface AssignComplaintRequest {
  assigned_to: string;
}

export interface UpdateStatusRequest {
  status: ComplaintStatus;
  resolution_note?: string;
}

export interface AddCommentRequest {
  body: string;
  is_internal: boolean;
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

export const listComplaints = (token: string, filter: ComplaintFilter = {}) =>
  apiFetch<ComplaintItem[]>(queryPath("complaints", filter), options(token));

export const listComplaintCategories = (token: string, activeOnly = true) =>
  apiFetch<ComplaintCategory[]>(`complaint-categories?active=${activeOnly}`, options(token));

export const getComplaintCategory = (token: string, id: string) =>
  apiFetch<ComplaintCategory>(`complaint-categories/${id}`, options(token));

export const createComplaintCategory = (token: string, data: CreateComplaintCategoryRequest) =>
  apiFetch<ComplaintCategory>("complaint-categories", options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateComplaintCategory = (token: string, id: string, data: UpdateComplaintCategoryRequest) =>
  apiFetch<ComplaintCategory>(`complaint-categories/${id}`, options(token, { method: "PATCH", body: JSON.stringify(data) }));

export const getComplaint = (token: string, id: string) =>
  apiFetch<ComplaintItem>(`complaints/${id}`, options(token));

export const createComplaint = (token: string, data: CreateComplaintRequest) =>
  apiFetch<ComplaintItem>("complaints", options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateComplaint = (token: string, id: string, data: UpdateComplaintRequest) =>
  apiFetch<ComplaintItem>(`complaints/${id}`, options(token, { method: "PATCH", body: JSON.stringify(data) }));

export const assignComplaint = (token: string, id: string, data: AssignComplaintRequest) =>
  apiFetch<ComplaintItem>(`complaints/${id}/assign`, options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateStatus = (token: string, id: string, data: UpdateStatusRequest) =>
  apiFetch<ComplaintItem>(`complaints/${id}/status`, options(token, { method: "POST", body: JSON.stringify(data) }));

export const addComment = (token: string, id: string, data: AddCommentRequest) =>
  apiFetch<CommentItem>(`complaints/${id}/comments`, options(token, { method: "POST", body: JSON.stringify(data) }));