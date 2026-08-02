import { apiFetch } from "./api";

export type AnnouncementCategory = "general" | "security" | "health" | "billing" | "event" | "emergency";
export type AnnouncementPriority = "normal" | "important";
export type AnnouncementStatus = "draft" | "scheduled" | "published" | "archived";
export type AnnouncementTargetType = "all" | "role" | "household" | "house_unit";
export type EventStatus = "planned" | "ongoing" | "completed" | "cancelled";

export interface TargetInput {
  target_type: AnnouncementTargetType;
  target_id?: string;
}

export interface AttachmentInfo {
  attachment_id: string;
  file_id: string;
  original_name: string;
  mime_type: string;
  size_bytes: number;
  purpose: string;
}

export interface AnnouncementTargetInfo extends TargetInput {
  target_name?: string;
}

export interface AnnouncementItem {
  id: string;
  title: string;
  content: string;
  category: AnnouncementCategory;
  priority: AnnouncementPriority;
  publish_at?: string;
  expire_at?: string;
  status: AnnouncementStatus;
  author_user_id: string;
  author_name: string;
  created_at: string;
  updated_at: string;
  is_read: boolean;
  read_count: number;
  targets: AnnouncementTargetInfo[];
  attachments: AttachmentInfo[];
}

export interface AnnouncementRequest {
  title: string;
  content: string;
  category: AnnouncementCategory;
  priority: AnnouncementPriority;
  publish_at?: string;
  expire_at?: string;
  status: Exclude<AnnouncementStatus, "archived">;
  targets: TargetInput[];
  attachment_file_ids: string[];
}

export interface ReadStats {
  total_audience: number;
  read_count: number;
}

export interface EventItem {
  id: string;
  title: string;
  description?: string;
  location?: string;
  starts_at: string;
  ends_at?: string;
  status: EventStatus;
  author_user_id: string;
  author_name: string;
  created_at: string;
  updated_at: string;
  attachments: AttachmentInfo[];
}

export interface EventRequest {
  title: string;
  description?: string;
  location?: string;
  starts_at: string;
  ends_at?: string;
  status: EventStatus;
  attachment_file_ids: string[];
}

export interface AnnouncementFilter {
  status?: AnnouncementStatus;
  category?: AnnouncementCategory;
  priority?: AnnouncementPriority;
  search?: string;
}

export interface EventFilter {
  status?: EventStatus;
  upcoming?: boolean;
  search?: string;
}

const options = (token: string, init: RequestInit = {}): RequestInit => ({
  ...init,
  headers: { ...init.headers, Authorization: `Bearer ${token}` },
});

const queryPath = <T extends object>(path: string, values: T) => {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== "" && value !== false) query.set(key, String(value));
  }
  return query.size ? `${path}?${query}` : path;
};

export const listAnnouncements = (token: string, filter: AnnouncementFilter = {}) =>
  apiFetch<AnnouncementItem[]>(queryPath("announcements", filter), options(token));

export const getAnnouncement = (token: string, id: string) =>
  apiFetch<AnnouncementItem>(`announcements/${id}`, options(token));

export const createAnnouncement = (token: string, data: AnnouncementRequest) =>
  apiFetch<AnnouncementItem>("announcements", options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateAnnouncement = (token: string, id: string, data: AnnouncementRequest) =>
  apiFetch<AnnouncementItem>(`announcements/${id}`, options(token, { method: "PATCH", body: JSON.stringify(data) }));

export const publishAnnouncement = (token: string, id: string) =>
  apiFetch<AnnouncementItem>(`announcements/${id}/publish`, options(token, { method: "POST" }));

export const archiveAnnouncement = (token: string, id: string) =>
  apiFetch<AnnouncementItem>(`announcements/${id}/archive`, options(token, { method: "POST" }));

export const getAnnouncementReadStats = (token: string, id: string) =>
  apiFetch<ReadStats>(`announcements/${id}/read-stats`, options(token));

export const listEvents = (token: string, filter: EventFilter = {}) =>
  apiFetch<EventItem[]>(queryPath("events", filter), options(token));

export const getEvent = (token: string, id: string) =>
  apiFetch<EventItem>(`events/${id}`, options(token));

export const createEvent = (token: string, data: EventRequest) =>
  apiFetch<EventItem>("events", options(token, { method: "POST", body: JSON.stringify(data) }));

export const updateEvent = (token: string, id: string, data: EventRequest) =>
  apiFetch<EventItem>(`events/${id}`, options(token, { method: "PATCH", body: JSON.stringify(data) }));

export const cancelEvent = (token: string, id: string) =>
  apiFetch<EventItem>(`events/${id}/cancel`, options(token, { method: "POST" }));