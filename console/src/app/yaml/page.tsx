"use client";
import { useState } from "react";
import { api } from "@/lib/api";

export default function YAMLPage() {
  const [yaml, setYaml] = useState("");
  const [status, setStatus] = useState<string | null>(null);

  return (
    <div className="space-y-6 lg:space-y-8">
      <header>
        <p className="label">YAML</p>
        <h1 className="mt-1 font-serif text-3xl text-ink-800 sm:text-4xl">YAML 进出</h1>
        <p className="mt-2 max-w-xl text-sm text-ink-500">
          与原版 OneBotFilter 配置文件兼容。导出当前所有 App 与 default 模板,或粘贴 YAML 进行批量导入。
        </p>
      </header>

      <section className="surface space-y-4 p-5 sm:p-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="font-medium text-ink-800">导出</h2>
          <a className="btn-clay self-start sm:self-auto" href={api.exportYAML()} download="filter.yaml">
            下载 filter.yaml
          </a>
        </div>
      </section>

      <section className="surface space-y-4 p-5 sm:p-6">
        <h2 className="font-medium text-ink-800">导入</h2>
        <textarea
          className="input min-h-[260px] font-mono text-xs"
          placeholder="将 OneBotFilter 风格的 YAML 粘贴在这里…"
          value={yaml}
          onChange={(e) => setYaml(e.target.value)}
        />
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <span className="text-xs text-ink-400">{status}</span>
          <button
            className="btn-primary self-start sm:self-auto"
            onClick={async () => {
              setStatus(null);
              const res = await api.importYAML(yaml);
              if (res.ok) {
                const data = await res.json();
                setStatus(`已导入 ${data.apps} 个 App`);
              } else {
                setStatus(`导入失败:${res.status} ${await res.text()}`);
              }
            }}
          >
            导入并重启
          </button>
        </div>
      </section>
    </div>
  );
}
