"use client";
import useSWR from "swr";
import { useState } from "react";
import { api, type FilterTemplate } from "@/lib/api";
import { RuleEditor } from "@/components/rule-editor";
import { Plus, Star, Trash2 } from "lucide-react";

export default function TemplatesPage() {
  const tpls = useSWR<FilterTemplate[]>("templates", api.listTemplates);
  const [editing, setEditing] = useState<FilterTemplate | null>(null);

  return (
    <div className="space-y-6 lg:space-y-8">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="label">规则模板</p>
          <h1 className="mt-1 font-serif text-3xl text-ink-800 sm:text-4xl">模板</h1>
          <p className="mt-2 max-w-xl text-sm text-ink-500">
            把一组规则封装成可复用的模板;每个下游 App 可以引用一个模板。模板「default」额外承担全局 ID 规则兜底。
          </p>
        </div>
        <button
          className="btn-clay self-start sm:self-auto"
          onClick={() =>
            setEditing({
              id: 0,
              name: "new-template",
              description: "",
              builtin: false,
              user_id_rules: "",
              group_id_rules: "",
              message_rules: "",
              private_message_rules: "",
              group_message_rules: "",
              created_at: "",
              updated_at: "",
            })
          }
        >
          <Plus size={14} />
          新建模板
        </button>
      </header>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {tpls.data?.map((t) => (
          <button
            key={t.id}
            onClick={() => setEditing(t)}
            className="surface group p-5 text-left transition-colors hover:border-clay-400"
          >
            <div className="flex items-center justify-between">
              <h3 className="flex items-center gap-2 font-serif text-xl text-ink-800">
                {t.name}
                {t.builtin ? <Star size={14} className="text-clay-500" /> : null}
              </h3>
              <span className="text-xs text-ink-400">#{t.id}</span>
            </div>
            <p className="mt-2 line-clamp-2 text-sm text-ink-500">
              {t.description || <span className="text-ink-300">(无描述)</span>}
            </p>
          </button>
        ))}
      </section>

      {editing ? (
        <TemplateEditor
          template={editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            tpls.mutate();
            setEditing(null);
          }}
        />
      ) : null}
    </div>
  );
}

function TemplateEditor({
  template,
  onClose,
  onSaved,
}: {
  template: FilterTemplate;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [draft, setDraft] = useState(template);
  const isNew = draft.id === 0;

  return (
    <div className="fixed inset-0 z-30 flex items-end justify-center bg-ink-900/30 sm:items-center sm:p-6">
      <div className="surface flex max-h-[92vh] w-full max-w-3xl flex-col overflow-hidden rounded-b-none sm:rounded-2xl">
        <header className="flex items-center justify-between gap-3 border-b border-ink-100 px-4 py-3 sm:px-6 sm:py-4">
          <div className="min-w-0">
            <p className="label">{isNew ? "新建模板" : "编辑模板"}</p>
            <h2 className="truncate font-serif text-xl text-ink-800 sm:text-2xl">
              {draft.name}
            </h2>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {!draft.builtin && !isNew ? (
              <button
                className="btn-ghost text-red-500"
                onClick={async () => {
                  if (confirm(`删除模板 ${draft.name}?`)) {
                    await api.deleteTemplate(draft.id);
                    onSaved();
                  }
                }}
              >
                <Trash2 size={14} /> 删除
              </button>
            ) : null}
            <button onClick={onClose} className="btn-ghost">
              关闭
            </button>
          </div>
        </header>
        <div className="flex-1 space-y-5 overflow-auto px-4 py-5 sm:px-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5">
              <span className="label">名称</span>
              <input
                className="input"
                value={draft.name}
                onChange={(e) => setDraft({ ...draft, name: e.target.value })}
                disabled={draft.builtin}
              />
            </label>
            <label className="flex flex-col gap-1.5">
              <span className="label">描述</span>
              <input
                className="input"
                value={draft.description}
                onChange={(e) => setDraft({ ...draft, description: e.target.value })}
              />
            </label>
          </div>
          <RuleEditor
            userIDRules={draft.user_id_rules}
            groupIDRules={draft.group_id_rules}
            messageRules={draft.message_rules}
            privateMessageRules={draft.private_message_rules}
            groupMessageRules={draft.group_message_rules}
            onChange={(p) => setDraft({ ...draft, ...p })}
          />
        </div>
        <footer className="flex items-center justify-end gap-2 border-t border-ink-100 bg-ink-50/60 px-4 py-3 sm:px-6 sm:py-4">
          <button className="btn-ghost" onClick={onClose}>
            取消
          </button>
          <button
            className="btn-primary"
            onClick={async () => {
              if (isNew) await api.createTemplate(draft);
              else await api.updateTemplate(draft.id, draft);
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
