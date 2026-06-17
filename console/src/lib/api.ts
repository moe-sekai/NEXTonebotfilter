export type IDRule = {
  mode: "" | "default" | "on" | "off" | "whitelist" | "blacklist";
  ids: number[];
};

export type MessageRule = {
  mode: "" | "default" | "on" | "off" | "whitelist" | "blacklist";
  filters: string[];
  prefix: string[];
  prefix_replace: string;
};

export type FilterGateway = {
  id: number;
  enabled: boolean;
  host: string;
  port: number;
  suffix: string;
  bot_id: string;
  access_token: string;
  user_agent: string;
  buffer_size: number;
  sleep_time: number;
  debug: boolean;
  dedup_enabled: boolean;
  dedup_ttl: number;
  updated_at: string;
};

export type FilterTemplate = {
  id: number;
  name: string;
  description: string;
  builtin: boolean;
  user_id_rules: string;
  group_id_rules: string;
  message_rules: string;
  private_message_rules: string;
  group_message_rules: string;
  created_at: string;
  updated_at: string;
};

export type FilterApp = {
  id: number;
  name: string;
  uri: string;
  access_token: string;
  enabled: boolean;
  builtin: boolean;
  internal: boolean;
  sort_order: number;
  template_id?: number | null;
  user_id_rules: string;
  group_id_rules: string;
  message_rules: string;
  private_message_rules: string;
  group_message_rules: string;
  created_at: string;
  updated_at: string;
};

export type Status = {
  running: boolean;
  listen: string;
  suffix: string;
  upstream_up: boolean;
  started_at?: string;
  upstreams: { self_id: string; remote: string; connected: boolean; since?: string }[];
  clients: { name: string; uri: string; connected: boolean; builtin: boolean }[];
};

export type FilterEvent = {
  seq: number;
  time: string;
  kind: "allow" | "block" | "prefix_pass" | "client_up" | "client_down" | "upstream_up" | "upstream_down";
  filter?: string;
  reason?: string;
  user_id?: number;
  group_id?: number;
  msg_type?: string;
  raw?: string;
};

const base = "/api";

// stripReadonly removes server-managed timestamp fields before sending a write
// request. The backend models use time.Time for created_at/updated_at, and Go's
// JSON decoder rejects empty-string values (e.g. on新建 forms), returning 400.
function stripReadonly<T extends object>(obj: T): Partial<T> {
  const copy = { ...obj } as Record<string, unknown>;
  delete copy.created_at;
  delete copy.updated_at;
  return copy as Partial<T>;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`${res.status} ${text}`);
  }
  if (res.status === 204) return undefined as T;
  const ct = res.headers.get("Content-Type") ?? "";
  if (ct.includes("json")) return (await res.json()) as T;
  return (await res.text()) as T;
}

export const api = {
  health: () => request<{ ok: boolean; time: string }>("/health"),
  status: () => request<Status>("/status"),
  restart: () => request<{ ok: boolean }>("/gateway/restart", { method: "POST" }),

  getGateway: () => request<FilterGateway>("/gateway"),
  saveGateway: (gw: FilterGateway) =>
    request<FilterGateway>("/gateway", { method: "PUT", body: JSON.stringify(stripReadonly(gw)) }),

  listApps: () => request<FilterApp[]>("/apps"),
  createApp: (a: Partial<FilterApp>) =>
    request<FilterApp>("/apps", { method: "POST", body: JSON.stringify(stripReadonly(a)) }),
  updateApp: (id: number, a: Partial<FilterApp>) =>
    request<FilterApp>(`/apps/${id}`, { method: "PUT", body: JSON.stringify(stripReadonly(a)) }),
  deleteApp: (id: number) => request<{ ok: boolean }>(`/apps/${id}`, { method: "DELETE" }),

  listTemplates: () => request<FilterTemplate[]>("/templates"),
  createTemplate: (t: Partial<FilterTemplate>) =>
    request<FilterTemplate>("/templates", { method: "POST", body: JSON.stringify(stripReadonly(t)) }),
  updateTemplate: (id: number, t: Partial<FilterTemplate>) =>
    request<FilterTemplate>(`/templates/${id}`, { method: "PUT", body: JSON.stringify(stripReadonly(t)) }),
  deleteTemplate: (id: number) =>
    request<{ ok: boolean }>(`/templates/${id}`, { method: "DELETE" }),

  recentEvents: (limit = 200) => request<FilterEvent[]>(`/events?limit=${limit}`),

  testRegex: (pattern: string, text: string) =>
    request<{ compiled: boolean; matched: boolean; error: string }>("/regex/test", {
      method: "POST",
      body: JSON.stringify({ pattern, text }),
    }),

  exportYAML: () => `${base}/yaml/export`,
  importYAML: (yaml: string) =>
    fetch(`${base}/yaml/import`, {
      method: "POST",
      headers: { "Content-Type": "application/x-yaml" },
      body: yaml,
    }),
};

export function decodeIDRule(raw: string): IDRule {
  if (!raw) return { mode: "", ids: [] };
  try {
    return JSON.parse(raw) as IDRule;
  } catch {
    return { mode: "", ids: [] };
  }
}

export function decodeMessageRule(raw: string): MessageRule {
  if (!raw) return { mode: "", filters: [], prefix: [], prefix_replace: "" };
  try {
    return JSON.parse(raw) as MessageRule;
  } catch {
    return { mode: "", filters: [], prefix: [], prefix_replace: "" };
  }
}

export function encodeIDRule(r: IDRule): string {
  return JSON.stringify({ ...r, ids: r.ids ?? [] });
}

export function encodeMessageRule(r: MessageRule): string {
  return JSON.stringify({
    mode: r.mode,
    filters: r.filters ?? [],
    prefix: r.prefix ?? [],
    prefix_replace: r.prefix_replace ?? "",
  });
}
