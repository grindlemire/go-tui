import { useState } from "react";
import { palette, useTheme } from "../lib/theme.ts";
import { loadLLMDoc } from "../lib/markdown.ts";

export default function SidebarLLMButton({ label }: { label: string }) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [copied, setCopied] = useState(false);

  return (
    <button
      onClick={() => {
        navigator.clipboard.writeText(loadLLMDoc());
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }}
      className="font-['Fira_Code',monospace] text-[10px] flex items-center gap-1.5 mt-6 px-3 py-1.5 rounded transition-all duration-200 w-full"
      style={{
        color: copied ? t.secondary : t.textDim,
        background: "transparent",
        border: "none",
        cursor: "pointer",
        borderTop: `1px solid ${t.border}`,
        paddingTop: "12px",
      }}
      onMouseEnter={(e) => {
        if (!copied) e.currentTarget.style.color = t.accent;
      }}
      onMouseLeave={(e) => {
        if (!copied) e.currentTarget.style.color = t.textDim;
      }}
      title="Copy all docs as a single LLM-optimized markdown file"
    >
      {copied ? (
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      ) : (
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      )}
      {copied ? "copied!" : label}
    </button>
  );
}
