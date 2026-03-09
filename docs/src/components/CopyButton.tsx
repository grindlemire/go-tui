import { useState } from "react";
import { palette, useTheme } from "../lib/theme.ts";

export default function CopyButton({
  text,
  label,
  copiedLabel,
  title,
}: {
  text: string;
  label: string;
  copiedLabel: string;
  title: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [copied, setCopied] = useState(false);

  return (
    <button
      onClick={() => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }}
      className="font-['Fira_Code',monospace] text-[11px] flex items-center gap-1.5 px-2.5 py-1 rounded transition-all duration-200"
      style={{
        color: copied ? t.secondary : t.textDim,
        background: "transparent",
        border: `1px solid ${copied ? t.secondary + "40" : "transparent"}`,
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        if (!copied) {
          e.currentTarget.style.color = t.accent;
          e.currentTarget.style.borderColor = t.accent + "30";
        }
      }}
      onMouseLeave={(e) => {
        if (!copied) {
          e.currentTarget.style.color = t.textDim;
          e.currentTarget.style.borderColor = "transparent";
        }
      }}
      title={title}
    >
      {copied ? (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      ) : (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      )}
      {copied ? copiedLabel : label}
    </button>
  );
}

export function RawMarkdownButton({ body }: { body: string }) {
  return (
    <div className="flex justify-end mb-3">
      <CopyButton
        text={body}
        label="raw markdown"
        copiedLabel="copied!"
        title="Copy raw markdown to clipboard"
      />
    </div>
  );
}
