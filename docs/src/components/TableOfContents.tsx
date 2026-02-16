import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import { useTheme, palette } from "../lib/theme.ts";
import { extractHeadingIds } from "./Markdown.tsx";

// Offset from top of viewport — must clear the sticky nav (h-12 = 48px) plus some breathing room
const SCROLL_OFFSET = 80;

export default function TableOfContents({ content }: { content: string }) {
  const { theme } = useTheme();
  const t = palette[theme];
  const entries = useMemo(
    () => extractHeadingIds(content)
      .filter((e) => e.level === 2)
      .map(({ text, id }) => ({ text, id })),
    [content],
  );
  const [activeId, setActiveId] = useState<string>("");
  const isScrollingTo = useRef(false);

  const updateActive = useCallback(() => {
    if (isScrollingTo.current) return;
    let current = "";
    for (const { id } of entries) {
      const el = document.getElementById(id);
      if (!el) continue;
      if (el.getBoundingClientRect().top <= SCROLL_OFFSET) {
        current = id;
      } else {
        break;
      }
    }
    // When at the top of the page (no heading past threshold), highlight first entry
    if (!current && entries.length > 0) {
      current = entries[0].id;
    }
    setActiveId(current);
  }, [entries]);

  useEffect(() => {
    if (entries.length === 0) return;
    updateActive();
    window.addEventListener("scroll", updateActive, { passive: true });
    return () => window.removeEventListener("scroll", updateActive);
  }, [updateActive]);

  if (entries.length === 0) return null;

  const scrollTo = (id: string) => {
    const el = document.getElementById(id);
    if (el) {
      isScrollingTo.current = true;
      setActiveId(id);
      history.replaceState(null, "", `#${id}`);
      const top = el.getBoundingClientRect().top + window.scrollY - SCROLL_OFFSET;
      window.scrollTo({ top: Math.max(0, top), behavior: "smooth" });
      // Release lock after scroll settles
      setTimeout(() => { isScrollingTo.current = false; }, 800);
    }
  };

  return (
    <nav className="w-44 shrink-0 hidden xl:block">
      <div className="sticky top-16 max-h-[calc(100vh-5rem)] overflow-y-auto custom-scroll">
        <div
          className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase mb-3"
          style={{ color: t.textDim }}
        >
          on this page
        </div>

        <div
          className="flex flex-col gap-0.5"
          style={{ borderLeft: `1px solid ${t.border}` }}
        >
          {entries.map((entry) => {
            const isActive = activeId === entry.id;
            return (
              <a
                key={entry.id}
                href={`#${entry.id}`}
                onClick={(e) => {
                  e.preventDefault();
                  scrollTo(entry.id);
                }}
                className="block font-['Fira_Code',monospace] transition-all duration-150 leading-snug"
                style={{
                  fontSize: "11px",
                  paddingLeft: 10,
                  paddingTop: 3,
                  paddingBottom: 3,
                  color: isActive ? t.accent : t.textDim,
                  borderLeft: `2px solid ${isActive ? t.accent : "transparent"}`,
                  marginLeft: -1,
                  textDecoration: "none",
                }}
                onMouseEnter={(e) => {
                  if (!isActive) e.currentTarget.style.color = t.textMuted;
                }}
                onMouseLeave={(e) => {
                  if (!isActive) e.currentTarget.style.color = t.textDim;
                }}
              >
                {entry.text}
              </a>
            );
          })}
        </div>
      </div>
    </nav>
  );
}
