"use client";
import { useEffect, useState } from "react";
import type { FilterEvent } from "@/lib/api";

export default function EventsPage() {
  const [events, setEvents] = useState<FilterEvent[]>([]);
  useEffect(() => {
    const es = new EventSource("/api/events/stream");
    es.onmessage = (e) => {
      try {
        const ev = JSON.parse(e.data) as FilterEvent;
        setEvents((prev) => [ev, ...prev].slice(0, 500));
      } catch {}
    };
    return () => es.close();
  }, []);

  const colors: Record<string, string> = {
    allow: "bg-emerald-500/10 text-emerald-600 border-emerald-500/30",
    block: "bg-red-500/10 text-red-600 border-red-500/30",
    prefix_pass: "bg-clay-500/10 text-clay-600 border-clay-500/30",
    client_up: "bg-sky-500/10 text-sky-600 border-sky-500/30",
    client_down: "bg-ink-200/40 text-ink-500 border-ink-200",
    upstream_up: "bg-violet-500/10 text-violet-600 border-violet-500/30",
    upstream_down: "bg-ink-200/40 text-ink-500 border-ink-200",
  };

  return (
    <div className="space-y-6">
      <header>
        <p className="label">实时事件</p>
        <h1 className="mt-1 font-serif text-3xl text-ink-800 sm:text-4xl">事件流</h1>
        <p className="mt-2 max-w-xl text-sm text-ink-500">
          通过 SSE 订阅网关的过滤事件:消息放行 / 拒绝 / 前缀替换,以及上下游连接状态变化。
        </p>
      </header>
      <div className="surface divide-y divide-ink-100">
        {events.length === 0 ? (
          <p className="py-10 text-center text-sm text-ink-400">等待事件…</p>
        ) : (
          events.map((ev) => (
            <div
              key={ev.seq}
              className="flex flex-wrap items-center gap-x-3 gap-y-1 px-4 py-3 text-sm sm:gap-4 sm:px-5"
            >
              <span className="w-16 shrink-0 font-mono text-xs text-ink-400 sm:w-20">
                #{ev.seq.toString().padStart(4, "0")}
              </span>
              <span className={`chip border ${colors[ev.kind] ?? "border-ink-200"}`}>{ev.kind}</span>
              <span className="font-medium text-ink-700">{ev.filter}</span>
              <span className="text-ink-400">{ev.reason}</span>
              <span className="ml-auto text-xs text-ink-300">
                {new Date(ev.time).toLocaleTimeString()}
              </span>
              {ev.raw ? (
                <code className="block w-full max-w-full truncate text-xs text-ink-500 sm:ml-2 sm:w-auto sm:max-w-md">
                  {ev.raw}
                </code>
              ) : null}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
