import { apiFetch } from "./api";

export interface EducationLevel {
  id: string;
  code: string;
  name: string;
  created_at: string;
}

export interface MaritalStatus {
  id: string;
  code: string;
  name: string;
  created_at: string;
}

const options = (token: string): RequestInit => ({
  headers: { Authorization: `Bearer ${token}` },
});

export const listEducationLevels = (token: string) =>
  apiFetch<EducationLevel[]>("education-levels", options(token));

export const listMaritalStatuses = (token: string) =>
  apiFetch<MaritalStatus[]>("marital-statuses", options(token));