// API client for /api/v1/me* privacy endpoints (PRIV-01/02/05).

export type ExportStatus =
  | "pending"
  | "running"
  | "ready"
  | "done"
  | "failed"
  | "expired";

export interface ExportJob {
  status: ExportStatus;
  bytes?: number;
  download_url?: string;
  expires_at?: string;
}

async function handle<T>(r: Response): Promise<T> {
  if (!r.ok) {
    const text = await r.text().catch(() => "");
    throw new Error(`HTTP ${r.status}: ${text}`);
  }
  if (r.status === 204) return undefined as unknown as T;
  return (await r.json()) as T;
}

function csrfHeader(): Record<string, string> {
  const m = document.cookie.match(/pesa_csrf=([^;]+)/);
  return m ? { "X-CSRF-Token": decodeURIComponent(m[1]) } : {};
}

export async function requestExport(): Promise<{ job_id: number }> {
  const r = await fetch("/api/v1/me/export", {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...csrfHeader() },
  });
  return handle(r);
}

export async function getExportJob(jobId: number): Promise<ExportJob> {
  const r = await fetch(`/api/v1/me/export/${jobId}`, {
    credentials: "same-origin",
  });
  return handle(r);
}

export function downloadExportUrl(token: string): string {
  return `/api/v1/me/export/download?t=${encodeURIComponent(token)}`;
}

export interface PatchEmailPayload {
  email: string;
  current_password: string;
}

export async function patchEmail(
  payload: PatchEmailPayload,
): Promise<{ message: string }> {
  const r = await fetch("/api/v1/me", {
    method: "PATCH",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...csrfHeader() },
    body: JSON.stringify(payload),
  });
  return handle(r);
}

export interface DeleteAccountPayload {
  current_password: string;
  confirmation_text: string;
  notes?: string;
}

export async function deleteAccount(
  payload: DeleteAccountPayload,
): Promise<void> {
  const r = await fetch("/api/v1/me", {
    method: "DELETE",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json", ...csrfHeader() },
    body: JSON.stringify(payload),
  });
  return handle(r);
}
