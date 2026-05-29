"use client";
import useSWR from "swr";
import { useState } from "react";
import { api, type FilterApp, type FilterTemplate } from "@/lib/api";
import { Plus, Trash2, ToggleLeft, ToggleRight } from "lucide-react";
import { RuleEditor } from "@/components/rule-editor";
import { Select } from "@/components/select";

export default function AppsPage() {
  const apps = useSWR("apps", api.listApps);
  const tpls = useSWR<FilterTemplate[]>("templates", api.listTemplates);

  const [editing, setEditing] = useState<FilterApp | null>(null);

  return (
    <div className="space-y-6 lg:space-y-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="label">下游应用</p>
          <h1 className="mt-1 font-serif text-3xl text-ink-800 sm:text-4xl">下游 App</h1>
          <p className="mt-2 max-w-xl text-sm text-ink-500">
            网关把上游 OneBot 客户端收到的事件分发给这里列出的下游 App,并在转发前应用各 App 的规则或所引用模板的规则。
          </p>
        </div>
        <button
          className="btn-clay self-start sm:self-auto"
          onClick={() => {
            const blank: FilterApp = {
              id: 0,
              name: "new-app",
              uri: "ws://127.0.0.1:8080",
              access_token: "",
              enabled: true,
              builtin: false,
              internal: false,
              sort_order: 0,
              template_id: tpls.data?.find((t) => t.name === "default")?.id ?? null,
              user_id_rules: "",
              group_id_rules: "",
              message_rules: "",
              private_message_rules: "",
              group_message_rules: "",
              created_at: "",
              updated_at: "",
            };
            setEditing(blank);
          }}
        >
          <Plus size={14} />
          新建 App
        </button>
      </header>

      {/* desktop table */}
      <section className="surface hidden overflow-hidden md:block">
        <table className="w-full text-sm">
          <thead className="bg-ink-50/60 text-left text-xs uppercase tracking-wider text-ink-400">
            <tr>
              <th className="px-5 py-3">名称</th>
              <th className="px-5 py-3">URI</th>
              <th className="px-5 py-3">模板</th>
              <th className="px-5 py-3">启用</th>
              <th className="px-5 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-ink-100">
            {apps.data?.map((app) => {
              const tpl = tpls.data?.find((t) => t.id === app.template_id);
              return (
                <tr key={app.id} className="hover:bg-ink-50/50">
                  <td className="px-5 py-3 font-medium text-ink-800">
                    <button onClick={() => setEditing(app)} className="hover:text-clay-500">
                      {app.name}
                    </button>
                    {app.builtin ? <span className="chip ml-2">built-in</span> : null}
                    {app.internal ? <span className="chip ml-2">internal</span> : null}
                  </td>
                  <td className="max-w-[280px] truncate px-5 py-3 font-mono text-xs text-ink-500">
                    {app.uri}
                  </td>
                  <td className="px-5 py-3 text-xs text-ink-500">{tpl?.name ?? "(自定义)"}</td>
                  <td className="px-5 py-3">
                    <button
                      onClick={async () => {
                        await api.updateApp(app.id, { ...app, enabled: !app.enabled });
                        apps.mutate();
                      }}
                      className="text-ink-500 hover:text-clay-500"
                    >
                      {app.enabled ? <ToggleRight size={20} /> : <ToggleLeft size={20} />}
                    </button>
                  </td>
                  <td className="px-5 py-3 text-right">
                    {!app.builtin ? (
                      <button
                        onClick={async () => {
                          if (confirm(`删除 ${app.name}?`)) {
                            await api.deleteApp(app.id);
                            apps.mutate();
                          }
                        }}
                        className="text-ink-400 hover:text-red-500"
                      >
                        <Trash2 size={14} />
                      </button>
                    ) : null}
                  </td>
                </tr>
              );
            })}
            {apps.data?.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-5 py-10 text-center text-sm text-ink-400">
                  还没有任何下游 App。
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </section>

      {/* mobile card list */}
      <section className="space-y-3 md:hidden">
        {apps.data?.length === 0 ? (
          <div className="surface p-6 text-center text-sm text-ink-400">
            还没有任何下游 App。
          </div>
        ) : null}
        {apps.data?.map((app) => {
          const tpl = tpls.data?.find((t) => t.id === app.template_id);
          return (
            <div key={app.id} className="surface p-4">
              <div className="flex items-start justify-between gap-3">
                <button
                  onClick={() => setEditing(app)}
                  className="flex min-w-0 flex-1 flex-col items-start text-left"
                >
                  <span className="flex flex-wrap items-center gap-2">
                    <span className="font-medium text-ink-800">{app.name}</span>
                    {app.builtin ? <span className="chip">built-in</span> : null}
                    {app.internal ? <span className="chip">internal</span> : null}
                  </span>
                  <code className="mt-1 block w-full truncate font-mono text-xs text-ink-500">
                    {app.uri}
                  </code>
                  <span className="mt-1 text-xs text-ink-400">
                    模板:{tpl?.name ?? "(自定义)"}
                  </span>
                </button>
                <div className="flex shrink-0 flex-col items-end gap-3">
                  <button
                    onClick={async () => {
                      await api.updateApp(app.id, { ...app, enabled: !app.enabled });
                      apps.mutate();
                    }}
                    className="text-ink-500 hover:text-clay-500"
                    aria-label={app.enabled ? "停用" : "启用"}
                  >
                    {app.enabled ? <ToggleRight size={22} /> : <ToggleLeft size={22} />}
                  </button>
                  {!app.builtin ? (
                    <button
                      onClick={async () => {
                        if (confirm(`删除 ${app.name}?`)) {
                          await api.deleteApp(app.id);
                          apps.mutate();
                        }
                      }}
                      className="text-ink-400 hover:text-red-500"
                      aria-label="删除"
                    >
                      <Trash2 size={16} />
                    </button>
                  ) : null}
                </div>
              </div>
            </div>
          );
        })}
      </section>

      {editing ? (
        <AppEditor
          app={editing}
          templates={tpls.data ?? []}
          onClose={() => setEditing(null)}
          onSaved={() => {
            apps.mutate();
            setEditing(null);
          }}
        />
      ) : null}
    </div>
  );
}

function AppEditor({
  app,
  templates,
  onClose,
  onSaved,
}: {
  app: FilterApp;
  templates: FilterTemplate[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const [draft, setDraft] = useState<FilterApp>(app);
  const isNew = draft.id === 0;
  const [tab, setTab] = useState<"basic" | "rules">("basic");

  return (
    <div className="fixed inset-0 z-30 flex items-end justify-center bg-ink-900/30 sm:items-center sm:p-6">
      <div className="surface flex max-h-[92vh] w-full max-w-3xl flex-col overflow-hidden rounded-b-none sm:rounded-2xl">
        <header className="flex items-center justify-between gap-3 border-b border-ink-100 px-4 py-3 sm:px-6 sm:py-4">
          <div className="min-w-0">
            <p className="label">{isNew ? "新建" : "编辑"}</p>
            <h2 className="truncate font-serif text-xl text-ink-800 sm:text-2xl">
              {draft.name}
            </h2>
          </div>
          <button onClick={onClose} className="btn-ghost shrink-0">
            关闭
          </button>
        </header>
        <div className="border-b border-ink-100 px-4 py-2 text-sm sm:px-6">
          {(["basic", "rules"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`mr-3 rounded-lg px-3 py-1 ${
                tab === t ? "bg-ink-100 text-ink-800" : "text-ink-500 hover:bg-ink-50"
              }`}
            >
              {t === "basic" ? "基础" : "规则"}
            </button>
          ))}
        </div>
        <div className="flex-1 overflow-auto px-4 py-5 sm:px-6">
          {tab === "basic" ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label="名称">
                <input
                  className="input"
                  value={draft.name}
                  onChange={(e) => setDraft({ ...draft, name: e.target.value })}
                />
              </Field>
              <Field label="URI">
                <input
                  className="input"
                  value={draft.uri}
                  onChange={(e) => setDraft({ ...draft, uri: e.target.value })}
                />
              </Field>
              <Field label="Access Token">
                <input
                  className="input"
                  value={draft.access_token}
                  onChange={(e) => setDraft({ ...draft, access_token: e.target.value })}
                />
              </Field>
              <Field label="模板">
                <Select<string>
                  value={draft.template_id ? String(draft.template_id) : ""}
                  onChange={(v) =>
                    setDraft({ ...draft, template_id: v ? Number(v) : null })
                  }
                  options={[
                    {
                      value: "",
                      label: "不使用模板",
                      description: "直接使用此 App 自身的规则字段",
                    },
                    ...templates.map((t) => ({
                      value: String(t.id),
                      label: t.name,
                      description: t.description || undefined,
                    })),
                  ]}
                />
              </Field>
              <Field label="排序">
                <input
                  type="number"
                  className="input"
                  value={draft.sort_order}
                  onChange={(e) => setDraft({ ...draft, sort_order: Number(e.target.value) })}
                />
              </Field>
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
            </div>
          ) : draft.template_id ? (
            <div className="surface-muted p-4 text-sm text-ink-500">
              此 App 引用模板 <span className="font-medium text-ink-700">
                {templates.find((t) => t.id === draft.template_id)?.name}
              </span>
              ,所有规则由模板提供。如需自定义,先在「基础」选项卡中将模板设为「不使用」。
            </div>
          ) : (
            <RuleEditor
              userIDRules={draft.user_id_rules}
              groupIDRules={draft.group_id_rules}
              messageRules={draft.message_rules}
              privateMessageRules={draft.private_message_rules}
              groupMessageRules={draft.group_message_rules}
              onChange={(p) => setDraft({ ...draft, ...p })}
            />
          )}
        </div>
        <footer className="flex items-center justify-end gap-2 border-t border-ink-100 bg-ink-50/60 px-4 py-3 sm:px-6 sm:py-4">
          <button className="btn-ghost" onClick={onClose}>
            取消
          </button>
          <button
            className="btn-primary"
            onClick={async () => {
              if (isNew) await api.createApp(draft);
              else await api.updateApp(draft.id, draft);
              onSaved();
            }}
          >
            保存
          </button>
        </footer>
      </div>
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
