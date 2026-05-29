"use client";
import useSWR from "swr";
import { useEffect, useState } from "react";
import { api, type FilterGateway } from "@/lib/api";

export default function GatewayPage() {
  const { data, mutate, isLoading } = useSWR<FilterGateway>("gateway", api.getGateway);
  const [draft, setDraft] = useState<FilterGateway | null>(null);
  useEffect(() => {
    if (data) setDraft(data);
  }, [data]);

  if (isLoading || !draft) {
    return <p className="text-sm text-ink-400">加载中…</p>;
  }

  return (
    <div className="space-y-6 lg:space-y-8">
      <header>
        <p className="label">网关</p>
        <h1 className="mt-1 font-serif text-3xl text-ink-800 sm:text-4xl">网关设置</h1>
        <p className="mt-2 max-w-xl text-sm text-ink-500">
          反向 WebSocket 监听参数,以及消息去重设置。保存后会重启网关。
        </p>
      </header>
      <section className="surface space-y-4 p-5 sm:p-6">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label="启用">
            <label className="flex items-center gap-2 text-sm text-ink-600">
              <input
                type="checkbox"
                checked={draft.enabled}
                onChange={(e) => setDraft({ ...draft, enabled: e.target.checked })}
              />
              Enabled
            </label>
          </Field>
          <Field label="Debug 日志">
            <label className="flex items-center gap-2 text-sm text-ink-600">
              <input
                type="checkbox"
                checked={draft.debug}
                onChange={(e) => setDraft({ ...draft, debug: e.target.checked })}
              />
              Debug
            </label>
          </Field>
          <Field label="Host">
            <input
              className="input"
              value={draft.host}
              onChange={(e) => setDraft({ ...draft, host: e.target.value })}
            />
          </Field>
          <Field label="Port">
            <input
              className="input"
              type="number"
              value={draft.port}
              onChange={(e) => setDraft({ ...draft, port: Number(e.target.value) })}
            />
          </Field>
          <Field label="Suffix">
            <input
              className="input"
              value={draft.suffix}
              onChange={(e) => setDraft({ ...draft, suffix: e.target.value })}
            />
          </Field>
          <Field label="Bot ID">
            <input
              className="input"
              value={draft.bot_id}
              onChange={(e) => setDraft({ ...draft, bot_id: e.target.value })}
            />
          </Field>
          <Field label="Access Token (上游)">
            <input
              className="input"
              value={draft.access_token}
              onChange={(e) => setDraft({ ...draft, access_token: e.target.value })}
            />
          </Field>
          <Field label="User-Agent">
            <input
              className="input"
              value={draft.user_agent}
              onChange={(e) => setDraft({ ...draft, user_agent: e.target.value })}
            />
          </Field>
          <Field label="Buffer Size">
            <input
              className="input"
              type="number"
              value={draft.buffer_size}
              onChange={(e) => setDraft({ ...draft, buffer_size: Number(e.target.value) })}
            />
          </Field>
          <Field label="Reconnect Sleep (s)">
            <input
              className="input"
              type="number"
              step="0.1"
              value={draft.sleep_time}
              onChange={(e) => setDraft({ ...draft, sleep_time: Number(e.target.value) })}
            />
          </Field>
          <Field label="去重">
            <label className="flex items-center gap-2 text-sm text-ink-600">
              <input
                type="checkbox"
                checked={draft.dedup_enabled}
                onChange={(e) => setDraft({ ...draft, dedup_enabled: e.target.checked })}
              />
              Dedup enabled
            </label>
          </Field>
          <Field label="Dedup TTL (s)">
            <input
              className="input"
              type="number"
              value={draft.dedup_ttl}
              onChange={(e) => setDraft({ ...draft, dedup_ttl: Number(e.target.value) })}
            />
          </Field>
        </div>
        <div className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end">
          <button className="btn-ghost" onClick={() => data && setDraft(data)}>
            重置
          </button>
          <button
            className="btn-primary"
            onClick={async () => {
              await api.saveGateway(draft);
              await mutate();
            }}
          >
            保存并重启
          </button>
        </div>
      </section>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="label">{label}</span>
      {children}
    </label>
  );
}
