"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { Activity, Boxes, Filter, Gauge, Menu, Settings, Sparkles, X } from "lucide-react";
import clsx from "clsx";

const items = [
  { href: "/", label: "概览", icon: Gauge },
  { href: "/apps", label: "下游 App", icon: Boxes },
  { href: "/templates", label: "规则模板", icon: Sparkles },
  { href: "/events", label: "实时事件", icon: Activity },
  { href: "/gateway", label: "网关设置", icon: Filter },
  { href: "/yaml", label: "YAML 进出", icon: Settings },
];

function NavBody({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  return (
    <>
      <div className="mb-10 flex items-center gap-2 px-2">
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-clay-500 text-white">
          <Filter size={18} />
        </div>
        <div className="flex flex-col">
          <span className="text-sm font-semibold text-ink-800">NEXTonebot</span>
          <span className="text-xs text-ink-400">filter console</span>
        </div>
      </div>
      <nav className="flex flex-col gap-1">
        {items.map(({ href, label, icon: Icon }) => {
          const active =
            href === "/" ? pathname === "/" : pathname?.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              onClick={onNavigate}
              className={clsx(
                "group flex items-center gap-3 rounded-xl px-3 py-2 text-sm transition-colors",
                active
                  ? "bg-white text-ink-800 shadow-soft"
                  : "text-ink-500 hover:bg-white/60 hover:text-ink-700",
              )}
            >
              <Icon
                size={16}
                className={clsx(active ? "text-clay-500" : "text-ink-400 group-hover:text-ink-600")}
              />
              <span>{label}</span>
            </Link>
          );
        })}
      </nav>
    </>
  );
}

export function Sidebar() {
  return (
    <aside className="hidden w-64 shrink-0 border-r border-ink-100 bg-ink-50/50 px-4 py-8 lg:block">
      <NavBody />
    </aside>
  );
}

export function MobileNav() {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  const current = items.find((i) =>
    i.href === "/" ? pathname === "/" : pathname?.startsWith(i.href),
  );

  // Close on route change.
  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  // Lock body scroll while drawer is open.
  useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = prev;
    };
  }, [open]);

  return (
    <>
      <header className="sticky top-0 z-30 flex items-center gap-3 border-b border-ink-100 bg-ink-50/95 px-4 py-3 backdrop-blur lg:hidden">
        <button
          aria-label="打开导航"
          className="flex h-9 w-9 items-center justify-center rounded-xl border border-ink-200 bg-white text-ink-700"
          onClick={() => setOpen(true)}
        >
          <Menu size={18} />
        </button>
        <div className="flex flex-1 items-center gap-2 truncate">
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-clay-500 text-white">
            <Filter size={14} />
          </div>
          <span className="truncate text-sm font-semibold text-ink-800">
            {current?.label ?? "NEXTonebot"}
          </span>
        </div>
      </header>
      {open ? (
        <>
          <div
            aria-hidden
            className="fixed inset-0 z-40 bg-ink-900/30 lg:hidden"
            onClick={() => setOpen(false)}
          />
          <aside className="fixed inset-y-0 left-0 z-50 w-72 max-w-[85vw] overflow-y-auto border-r border-ink-100 bg-ink-50 px-4 py-6 shadow-xl lg:hidden">
            <div className="mb-4 flex items-center justify-between px-2">
              <span className="label">导航</span>
              <button
                aria-label="关闭"
                onClick={() => setOpen(false)}
                className="flex h-8 w-8 items-center justify-center rounded-lg text-ink-500 hover:bg-ink-100"
              >
                <X size={16} />
              </button>
            </div>
            <NavBody onNavigate={() => setOpen(false)} />
          </aside>
        </>
      ) : null}
    </>
  );
}
