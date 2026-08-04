import { apiFetch } from "./api";

export interface OrganizationSettings {
  id: string;
  name: string;
  rt_number: string;
  rw_number: string;
  address: string | null;
  timezone: string;
  logo_file_id: string | null;
  bank_name: string | null;
  bank_account_number: string | null;
  bank_account_holder: string | null;
  max_upload_size_bytes: number;
  default_letter_number_pattern: string;
  status: string;
  settings: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface UpdateOrganizationSettingsRequest {
  name?: string;
  rt_number?: string;
  rw_number?: string;
  address?: string | null;
  timezone?: string;
  logo_file_id?: string | null;
  bank_name?: string | null;
  bank_account_number?: string | null;
  bank_account_holder?: string | null;
  max_upload_size_bytes?: number;
  default_letter_number_pattern?: string;
  settings?: Record<string, unknown>;
}

export interface AuditLogItem {
  id: number;
  actor_user_id: string | null;
  actor_name: string | null;
  actor_role_codes: string[];
  action: string;
  entity_type: string;
  entity_id: string | null;
  metadata: Record<string, unknown>;
  before_data: Record<string, unknown>;
  after_data: Record<string, unknown>;
  request_id: string | null;
  created_at: string;
}

export interface AuditLogListResult {
  data: AuditLogItem[];
  meta: {
    next_cursor?: number;
    has_more: boolean;
  };
}

export async function getOrganizationSettings(
  accessToken: string,
): Promise<OrganizationSettings> {
  return apiFetch<OrganizationSettings>("organizations/current", {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}

export async function updateOrganizationSettings(
  accessToken: string,
  data: UpdateOrganizationSettingsRequest,
): Promise<OrganizationSettings> {
  return apiFetch<OrganizationSettings>("organizations/current", {
    method: "PATCH",
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(data),
  });
}

export async function getAuditLogs(
  accessToken: string,
  params?: {
    action?: string;
    actor_user_id?: string;
    entity_type?: string;
    entity_id?: string;
    limit?: number;
    cursor?: number;
  },
): Promise<AuditLogListResult> {
  const search = new URLSearchParams();
  if (params?.action) search.set("action", params.action);
  if (params?.actor_user_id) search.set("actor_user_id", params.actor_user_id);
  if (params?.entity_type) search.set("entity_type", params.entity_type);
  if (params?.entity_id) search.set("entity_id", params.entity_id);
  if (params?.limit) search.set("limit", params.limit.toString());
  if (params?.cursor) search.set("cursor", params.cursor.toString());

  const query = search.toString();
  return apiFetch<AuditLogListResult>(
    `audit-logs${query ? `?${query}` : ""}`,
    {
      headers: { Authorization: `Bearer ${accessToken}` },
    },
    false,
  );
}

export async function getAuditLog(
  accessToken: string,
  id: number,
): Promise<AuditLogItem> {
  return apiFetch<AuditLogItem>(`audit-logs/${id}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });
}