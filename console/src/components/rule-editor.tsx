"use client";
import {
  decodeIDRule,
  decodeMessageRule,
  encodeIDRule,
  encodeMessageRule,
  type IDRule,
  type MessageRule,
} from "@/lib/api";
import { useEffect, useMemo, useRef, useState } from "react";
import { Select, type SelectOption } from "@/components/select";

type Props = {
  userIDRules: string;
  groupIDRules: string;
  messageRules: string;
  privateMessageRules: string;
  groupMessageRules: string;
  onChange: (patch: {
    user_id_rules?: string;
    group_id_rules?: string;
    message_rules?: string;
    private_message_rules?: string;
    group_message_rules?: string;
  }) => void;
};

const ID_OPTIONS: SelectOption<IDRule["mode"]>[] = [
  { value: "", label: "未设置", description: "等同于 default,沿用上层兜底规则" },
  { value: "default", label: "继承默认", description: "使用 default 模板里的全局兜底" },
  { value: "on", label: "全部放行", description: "任意 ID 均通过此规则" },
  { value: "off", label: "全部拦截", description: "任意 ID 均被拦截" },
  {
    value: "whitelist",
    label: "白名单",
    description: "仅下方 ID 列表中的对象通过,其它一律拦截",
  },
  {
    value: "blacklist",
    label: "黑名单",
    description: "下方 ID 列表中的对象被拦截,其它放行",
  },
];

const MSG_OPTIONS: SelectOption<MessageRule["mode"]>[] = [
  { value: "", label: "未设置", description: "等同于 default,沿用通用消息规则" },
  { value: "default", label: "继承通用", description: "使用上方“通用消息规则”的设置" },
  { value: "on", label: "全部放行", description: "所有消息直接通过,不做正则匹配" },
  { value: "off", label: "全部拦截", description: "所有消息一律拦截" },
  {
    value: "whitelist",
    label: "正则白名单",
    description: "仅命中下方任一正则的消息放行,其余拦截",
  },
  {
    value: "blacklist",
    label: "正则黑名单",
    description: "命中任一正则的消息被拦截,其余放行",
  },
];

const ID_LIST_VISIBLE: ReadonlySet<IDRule["mode"]> = new Set(["whitelist", "blacklist"]);
const MSG_PATTERNS_VISIBLE: ReadonlySet<MessageRule["mode"]> = new Set([
  "whitelist",
  "blacklist",
]);

export function RuleEditor(props: Props) {
  const userID = useMemo(() => decodeIDRule(props.userIDRules), [props.userIDRules]);
  const groupID = useMemo(() => decodeIDRule(props.groupIDRules), [props.groupIDRules]);
  const message = useMemo(() => decodeMessageRule(props.messageRules), [props.messageRules]);
  const priv = useMemo(() => decodeMessageRule(props.privateMessageRules), [props.privateMessageRules]);
  const grp = useMemo(() => decodeMessageRule(props.groupMessageRules), [props.groupMessageRules]);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <IDRuleBlock
          title="用户 ID 规则"
          subtitle="按 user_id 控制私聊与群聊中单个用户的可见性"
          rule={userID}
          onChange={(r) => props.onChange({ user_id_rules: encodeIDRule(r) })}
        />
        <IDRuleBlock
          title="群 ID 规则"
          subtitle="按 group_id 控制哪些群的消息会被转发"
          rule={groupID}
          onChange={(r) => props.onChange({ group_id_rules: encodeIDRule(r) })}
        />
      </div>
      <MessageRuleBlock
        title="通用消息规则"
        subtitle="作为私聊 / 群聊消息的默认行为;下方两块若选“继承通用”就走这里"
        rule={message}
        onChange={(r) => props.onChange({ message_rules: encodeMessageRule(r) })}
      />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <MessageRuleBlock
          title="私聊消息规则"
          subtitle="只对 message_type = private 生效;留“继承通用”即沿用上方设置"
          rule={priv}
          onChange={(r) => props.onChange({ private_message_rules: encodeMessageRule(r) })}
        />
        <MessageRuleBlock
          title="群聊消息规则"
          subtitle="只对 message_type = group 生效;留“继承通用”即沿用上方设置"
          rule={grp}
          onChange={(r) => props.onChange({ group_message_rules: encodeMessageRule(r) })}
        />
      </div>
    </div>
  );
}

function IDRuleBlock({
  title,
  subtitle,
  rule,
  onChange,
}: {
  title: string;
  subtitle: string;
  rule: IDRule;
  onChange: (r: IDRule) => void;
}) {
  const showList = ID_LIST_VISIBLE.has(rule.mode);
  return (
    <div className="surface-muted space-y-3 p-4">
      <div className="space-y-1">
        <span className="label">{title}</span>
        <p className="text-xs leading-snug text-ink-400">{subtitle}</p>
      </div>
      <Select<IDRule["mode"]>
        value={rule.mode}
        onChange={(m) => onChange({ ...rule, mode: m })}
        options={ID_OPTIONS}
      />
      <div className="space-y-1.5">
        <span className="text-[11px] uppercase tracking-wider text-ink-400">
          ID 列表(每行一个,也可以用逗号分隔)
        </span>
        <IDListTextarea
          ids={rule.ids ?? []}
          disabled={!showList}
          placeholder={
            showList
              ? "10001\n20002\n30003"
              : `当前模式“${labelOf(ID_OPTIONS, rule.mode)}”不需要 ID 列表`
          }
          onChange={(ids) => onChange({ ...rule, ids })}
        />
        {!showList ? (
          <p className="text-[11px] text-ink-400">
            仅在“白名单 / 黑名单”模式下需要填写 ID,其它模式留空即可。
          </p>
        ) : null}
      </div>
    </div>
  );
}

function IDListTextarea({
  ids,
  onChange,
  disabled,
  placeholder,
}: {
  ids: number[];
  onChange: (next: number[]) => void;
  disabled: boolean;
  placeholder: string;
}) {
  // The textarea owns its raw text so newlines, trailing spaces and partial
  // edits aren't clobbered by the parsed-numbers round-trip. We only resync
  // from `ids` when the parent hands us a value that wasn't produced by our
  // last emission (e.g. switching to a different template / app).
  const [text, setText] = useState<string>(() => idsToText(ids));
  const lastEmittedRef = useRef<number[]>(ids);

  useEffect(() => {
    if (!sameIDs(lastEmittedRef.current, ids)) {
      setText(idsToText(ids));
      lastEmittedRef.current = ids;
    }
  }, [ids]);

  return (
    <textarea
      className="input min-h-[80px] font-mono text-xs disabled:cursor-not-allowed disabled:bg-ink-50/60 disabled:text-ink-300"
      placeholder={placeholder}
      value={text}
      disabled={disabled}
      spellCheck={false}
      onChange={(e) => {
        const v = e.target.value;
        setText(v);
        const parsed = v
          .split(/[\s,]+/)
          .map((s) => Number(s.trim()))
          .filter((n) => !Number.isNaN(n) && n !== 0);
        lastEmittedRef.current = parsed;
        onChange(parsed);
      }}
    />
  );
}

function idsToText(ids: number[]): string {
  return (ids ?? []).join("\n");
}

function sameIDs(a: number[], b: number[]): boolean {
  if (a === b) return true;
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

function MessageRuleBlock({
  title,
  subtitle,
  rule,
  onChange,
}: {
  title: string;
  subtitle: string;
  rule: MessageRule;
  onChange: (r: MessageRule) => void;
}) {
  const showPatterns = MSG_PATTERNS_VISIBLE.has(rule.mode);
  // prefix 始终可用:它是“以特定前缀开头的消息直接放行”,不依赖 mode。
  const prefixDisabled = rule.mode === "off";
  return (
    <div className="surface-muted space-y-4 p-4">
      <div className="space-y-1">
        <span className="label">{title}</span>
        <p className="text-xs leading-snug text-ink-400">{subtitle}</p>
      </div>
      <Select<MessageRule["mode"]>
        value={rule.mode}
        onChange={(m) => onChange({ ...rule, mode: m })}
        options={MSG_OPTIONS}
      />
      <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
        <Field
          label="正则规则(每行一个)"
          hint={
            showPatterns
              ? "命中任意一条即视为匹配,引擎为 .NET 风格的 regexp2"
              : `当前模式“${labelOf(MSG_OPTIONS, rule.mode)}”不使用正则`
          }
        >
          <textarea
            className="input min-h-[80px] font-mono text-xs disabled:cursor-not-allowed disabled:bg-ink-50/60 disabled:text-ink-300"
            value={(rule.filters ?? []).join("\n")}
            disabled={!showPatterns}
            spellCheck={false}
            onChange={(e) =>
              onChange({ ...rule, filters: e.target.value.split("\n") })
            }
          />
        </Field>
        <Field
          label="消息前缀(每行一个)"
          hint={
            prefixDisabled
              ? "“全部拦截”模式下不允许任何消息透传"
              : "命中任一前缀的消息直接放行,可同时与白/黑名单共存"
          }
        >
          <textarea
            className="input min-h-[80px] font-mono text-xs disabled:cursor-not-allowed disabled:bg-ink-50/60 disabled:text-ink-300"
            value={(rule.prefix ?? []).join("\n")}
            disabled={prefixDisabled}
            spellCheck={false}
            onChange={(e) =>
              onChange({ ...rule, prefix: e.target.value.split("\n") })
            }
          />
        </Field>
      </div>
      <Field
        label="前缀替换"
        hint={
          prefixDisabled
            ? "“全部拦截”模式下无效"
            : "把命中的前缀替换成此字符串后再转发,留空即去掉前缀"
        }
      >
        <input
          className="input disabled:cursor-not-allowed disabled:bg-ink-50/60 disabled:text-ink-300"
          placeholder="例如:/  或留空"
          value={rule.prefix_replace ?? ""}
          disabled={prefixDisabled}
          onChange={(e) => onChange({ ...rule, prefix_replace: e.target.value })}
        />
      </Field>
    </div>
  );
}

function Field({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[11px] uppercase tracking-wider text-ink-400">{label}</span>
      {children}
      {hint ? <span className="text-[11px] leading-snug text-ink-400">{hint}</span> : null}
    </label>
  );
}

function labelOf<T extends string>(options: SelectOption<T>[], value: T): string {
  return options.find((o) => o.value === value)?.label ?? value;
}
