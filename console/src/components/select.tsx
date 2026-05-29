"use client";
import { useEffect, useRef, useState } from "react";
import { Check, ChevronDown } from "lucide-react";
import clsx from "clsx";

export type SelectOption<T extends string = string> = {
  value: T;
  label: string;
  description?: string;
};

type Props<T extends string> = {
  value: T;
  onChange: (v: T) => void;
  options: SelectOption<T>[];
  className?: string;
  placeholder?: string;
  disabled?: boolean;
};

export function Select<T extends string>({
  value,
  onChange,
  options,
  className,
  placeholder = "请选择",
  disabled,
}: Props<T>) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const current = options.find((o) => o.value === value);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    const esc = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    document.addEventListener("keydown", esc);
    return () => {
      document.removeEventListener("mousedown", handler);
      document.removeEventListener("keydown", esc);
    };
  }, [open]);

  return (
    <div ref={ref} className={clsx("relative", className)}>
      <button
        type="button"
        disabled={disabled}
        onClick={() => setOpen((o) => !o)}
        className={clsx(
          "flex w-full items-center justify-between gap-2 rounded-xl border px-3 py-2 text-left text-sm transition-colors",
          disabled
            ? "cursor-not-allowed border-ink-100 bg-ink-50/60 text-ink-300"
            : "border-ink-200 bg-white text-ink-800 hover:border-ink-300 focus:border-ink-500 focus:outline-none focus:ring-2 focus:ring-ink-200",
          open && !disabled && "border-ink-500 ring-2 ring-ink-200",
        )}
      >
        <span className={current ? "" : "text-ink-300"}>
          {current?.label ?? placeholder}
        </span>
        <ChevronDown
          size={14}
          className={clsx(
            "shrink-0 text-ink-400 transition-transform",
            open && "rotate-180",
          )}
        />
      </button>
      {open && !disabled ? (
        <div className="absolute left-0 right-0 z-50 mt-1.5 overflow-hidden rounded-xl border border-ink-200 bg-white shadow-lg">
          <ul className="max-h-72 overflow-auto py-1" role="listbox">
            {options.map((opt) => {
              const active = opt.value === value;
              return (
                <li key={opt.value || "__empty"}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={active}
                    onClick={() => {
                      onChange(opt.value);
                      setOpen(false);
                    }}
                    className={clsx(
                      "flex w-full items-start gap-3 px-3 py-2 text-left text-sm transition-colors",
                      active ? "bg-ink-50" : "hover:bg-ink-50/60",
                    )}
                  >
                    <Check
                      size={14}
                      className={clsx(
                        "mt-0.5 shrink-0",
                        active ? "text-clay-500" : "opacity-0",
                      )}
                    />
                    <span className="flex flex-1 flex-col gap-0.5">
                      <span className="font-medium text-ink-800">{opt.label}</span>
                      {opt.description ? (
                        <span className="text-xs leading-snug text-ink-400">
                          {opt.description}
                        </span>
                      ) : null}
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
        </div>
      ) : null}
    </div>
  );
}
