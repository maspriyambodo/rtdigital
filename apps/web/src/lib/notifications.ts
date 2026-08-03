import { apiFetch } from "./api";

export interface Notification {
  id: string;
  type: string;
  title: string;
  body?: string;
  reference_type?: string;
  reference_id?: string;
  read_at?: string;
  created_at: string;
}

export async function fetchNotifications(
  accessToken: string,
  unreadOnly = false,
  limit = 50,
): Promise<Notification[]> {
  const query = new URLSearchParams({
    unread_only: String(unreadOnly),
    limit: String(limit),
  });

  return apiFetch<Notification[]>(`notifications?${query}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function markNotificationAsRead(
  accessToken: string,
  notificationID: string,
): Promise<Notification> {
  return apiFetch<Notification>(`notifications/${notificationID}/read`, {
    method: "PATCH",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function markAllNotificationsAsRead(
  accessToken: string,
): Promise<{ updated_count: number }> {
  return apiFetch<{ updated_count: number }>("notifications/read-all", {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}