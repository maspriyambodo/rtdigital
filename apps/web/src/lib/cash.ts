import { apiFetch } from "./api";

export type CashType = "income" | "expense";

export interface CashCategory {
  id: string;
  name: string;
  type: CashType;
  status: "active" | "inactive";
  created_at: string;
  updated_at: string;
}

export interface CashTransaction {
  id: string;
  transaction_number: string;
  type: CashType;
  category_id?: string;
  category_name?: string;
  amount: number;
  transaction_date: string;
  description: string;
  proof_file_id?: string;
  reference_type?: string;
  reference_id?: string;
  reversal_of_id?: string;
  status: "active" | "reversed";
  created_by: string;
  created_at: string;
  updated_at: string;
  running_balance: number;
}

export interface CashBook {
  transactions: CashTransaction[];
  total_income: number;
  total_expense: number;
  balance: number;
}

export interface CashBookFilter {
  startDate?: string;
  endDate?: string;
  type?: CashType;
  categoryId?: string;
}

export async function getCashBook(token: string, filter: CashBookFilter = {}): Promise<CashBook> {
  const params = new URLSearchParams();
  if (filter.startDate) params.set("start_date", filter.startDate);
  if (filter.endDate) params.set("end_date", filter.endDate);
  if (filter.type) params.set("type", filter.type);
  if (filter.categoryId) params.set("category_id", filter.categoryId);
  const query = params.toString();
  return apiFetch<CashBook>(`cash/book${query ? `?${query}` : ""}`, {}, true, token);
}

export function getCashCategories(token: string): Promise<CashCategory[]> {
  return apiFetch<CashCategory[]>("cash/categories", {}, true, token);
}

export function createCashCategory(
  token: string,
  data: { name: string; type: CashType },
): Promise<CashCategory> {
  return apiFetch<CashCategory>(
    "cash/categories",
    { method: "POST", body: JSON.stringify(data) },
    true,
    token,
  );
}

export function updateCashCategory(
  token: string,
  id: string,
  data: { name?: string; status?: "active" | "inactive" },
): Promise<CashCategory> {
  return apiFetch<CashCategory>(
    `cash/categories/${id}`,
    { method: "PATCH", body: JSON.stringify(data) },
    true,
    token,
  );
}

export function recordCashTransaction(
  token: string,
  data: {
    type: CashType;
    categoryId: string;
    amount: number;
    transactionDate: string;
    description: string;
    proofFileId?: string;
  },
): Promise<CashTransaction> {
  return apiFetch<CashTransaction>(
    "cash/transactions",
    {
      method: "POST",
      body: JSON.stringify({
        type: data.type,
        category_id: data.categoryId,
        amount: data.amount,
        transaction_date: data.transactionDate,
        description: data.description,
        proof_file_id: data.proofFileId,
      }),
    },
    true,
    token,
  );
}

export function reverseCashTransaction(
  token: string,
  id: string,
  reason: string,
): Promise<CashTransaction> {
  return apiFetch<CashTransaction>(
    `cash/transactions/${id}/reverse`,
    { method: "POST", body: JSON.stringify({ reason }) },
    true,
    token,
  );
}
