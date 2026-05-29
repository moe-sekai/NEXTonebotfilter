"use client";
import useSWR from "swr";
import { api, type Status } from "@/lib/api";
import { Activity, ArrowUpRight, Cable, RefreshCw, Wifi } from "lucide-react";

const fetcher = () => api.status();

export default function HomePage() {
  const { data, isLoading, mutate } = useSWR<Status>("status", fetcher, {
    refreshInterval: 4000,
  });

  return (
    <div className="space-y-8 lg:space-y-10">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="label">概览</p>
          <h1 className="mt-1 font-serif text-3xl text-ink-800 sm:text-4xl">网关状态</h1>
          <p className="mt-2 max-w-xl text-sm text-ink-500">
            一份独立运行的 OneBot 反向 WS 过滤网关。规则模板在左侧管理,这里展示运行时实况。
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => mutate()} className="btn-outline">
            <RefreshCw size={14} />
            刷新
          </button>
          <button onClick={() => api.restart().then(() => mutate())} className="btn-clay">
            重启网关
          </button>
        </div>
      </header>

      <section className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
        <Card
          icon={<Wifi size={16} />}
          title="网关"
          value={isLoading ? "—" : data?.running ? "运行中" : "已停止"}
          hint={data?.listen ? `监听 ${data.listen}${data.suffix}` : "未启动"}
          accent={data?.running}
        />
        <Card
          icon={<Cable size={16} />}
          title="上游 OneBot"
          value={
            isLoading
              ? "—"
              : data?.upstream_up
                ? `${data.upstreams.length} 个连接`
                : "无连接"
          }
          hint={data?.upstreams?.[0]?.remote ?? "等待 OneBot 客户端反连"}
          accent={data?.upstream_up}
        />
        <Card
          icon={<Activity size={16} />}
          title="下游客户端"
          value={isLoading ? "—" : `${data?.clients?.filter((c) => c.connected).length ?? 0}/${data?.clients?.length ?? 0}`}
          hint="已连接 / 总数"
        />
      </section>

      <section className="surface p-5 sm:p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-base font-semibold text-ink-800">下游应用</h2>
          <a href="/apps" className="text-xs text-clay-500 hover:text-clay-600">
            管理 <ArrowUpRight size={12} className="inline" />
          </a>
        </div>
        {data?.clients?.length ? (
          <ul className="divide-y divide-ink-100">
            {data.clients.map((c) => (
              <li
                key={c.name}
                className="flex flex-col gap-2 py-3 text-sm sm:flex-row sm:items-center sm:justify-between sm:gap-4"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <span
                    className={`h-2 w-2 shrink-0 rounded-full ${c.connected ? "bg-emerald-500" : "bg-ink-300"}`}
                  />
                  <span className="truncate font-medium text-ink-800">{c.name}</span>
                  {c.builtin ? <span className="chip">built-in</span> : null}
                </div>
                <code className="truncate text-xs text-ink-400 sm:text-right">{c.uri}</code>
              </li>
            ))}
          </ul>
        ) : (
          <p className="py-6 text-center text-sm text-ink-400">尚无下游 App。前往「下游 App」添加第一个。</p>
        )}
      </section>
    </div>
  );
}

function Card({
  icon,
  title,
  value,
  hint,
  accent,
}: {
  icon: React.ReactNode;
  title: string;
  value: string;
  hint?: string;
  accent?: boolean;
}) {
  return (
    <div className="surface p-5">
      <div className="flex items-center justify-between text-ink-400">
        <span className="label flex items-center gap-2">
          {icon}
          {title}
        </span>
      </div>
      <div className="mt-3 flex items-baseline gap-2">
        <span
          className={`font-serif text-3xl ${accent ? "text-clay-500" : "text-ink-800"}`}
        >
          {value}
        </span>
      </div>
      {hint ? <p className="mt-1 text-xs text-ink-400">{hint}</p> : null}
    </div>
  );
}
