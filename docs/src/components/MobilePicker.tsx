import { useState } from "react";
import { palette, useTheme } from "../lib/theme.ts";

export default function MobilePicker({
  pages,
  activeIndex,
  onSelect,
}: {
  pages: { slug: string; title: string }[];
  activeIndex: number;
  onSelect: (index: number) => void;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [open, setOpen] = useState(false);

  return (
    <div className="md:hidden mb-6 sm:mb-8 relative">
      <button
        onClick={() => setOpen(!open)}
        className="font-['Fira_Code',monospace] w-full rounded-lg px-4 py-3 text-[13px] text-left flex items-center justify-between gap-2"
        style={{
          background: t.bgSecondary,
          color: t.text,
          border: `1px solid ${t.border}`,
        }}
      >
        <div className="flex items-center gap-2.5 min-w-0">
          <span
            className="text-[10px] px-1.5 py-0.5 rounded shrink-0"
            style={{
              background: `${t.accent}15`,
              color: t.accent,
              border: `1px solid ${t.accent}30`,
            }}
          >
            {String(activeIndex + 1).padStart(2, "0")}
          </span>
          <span className="truncate">{pages[activeIndex].title}</span>
        </div>
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          className="shrink-0 transition-transform duration-200"
          style={{
            transform: open ? "rotate(180deg)" : "rotate(0deg)",
            color: t.textDim,
          }}
        >
          <path d="M2.5 4.5L6 8L9.5 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-30" onClick={() => setOpen(false)} />
          <div
            className="absolute top-full left-0 right-0 mt-1 rounded-lg overflow-hidden z-40"
            style={{
              background: t.bgSecondary,
              border: `1px solid ${t.border}`,
              boxShadow: theme === "dark"
                ? "0 8px 24px rgba(0,0,0,0.5)"
                : "0 8px 24px rgba(0,0,0,0.12)",
            }}
          >
            {pages.map((page, i) => {
              const active = i === activeIndex;
              return (
                <button
                  key={page.slug}
                  onClick={() => {
                    onSelect(i);
                    setOpen(false);
                  }}
                  className="font-['Fira_Code',monospace] w-full text-left px-4 py-2.5 text-[12px] flex items-center gap-2.5 transition-colors duration-100"
                  style={{
                    color: active ? t.accent : t.textMuted,
                    background: active
                      ? `${t.accent}0a`
                      : "transparent",
                    borderBottom: i < pages.length - 1 ? `1px solid ${t.border}` : "none",
                  }}
                >
                  <span
                    className="text-[10px] w-5 text-center shrink-0"
                    style={{ color: active ? t.accent : t.textDim }}
                  >
                    {String(i + 1).padStart(2, "0")}
                  </span>
                  {page.title}
                </button>
              );
            })}
          </div>
        </>
      )}
    </div>
  );
}
