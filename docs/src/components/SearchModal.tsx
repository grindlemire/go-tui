import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useTheme, palette } from "../lib/theme.ts";
import { search, type SearchHit } from "../lib/search.ts";

interface SearchModalProps {
  open: boolean;
  onClose: () => void;
}

export default function SearchModal({ open, onClose }: SearchModalProps) {
  const { theme } = useTheme();
  const t = palette[theme];
  const navigate = useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);
  const modalRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const [results, setResults] = useState<SearchHit[]>([]);
  const [closing, setClosing] = useState(false);

  // Reset state when opening
  useEffect(() => {
    if (open) {
      setQuery("");
      setResults([]);
      setSelected(0);
      setClosing(false);
      // Focus input after mount
      requestAnimationFrame(() => inputRef.current?.focus());
    }
  }, [open]);

  // Search on query change
  useEffect(() => {
    const hits = search(query);
    setResults(hits);
    setSelected(0);
  }, [query]);

  // Scroll selected item into view
  useEffect(() => {
    if (!listRef.current) return;
    const item = listRef.current.children[selected] as HTMLElement | undefined;
    item?.scrollIntoView({ block: "nearest" });
  }, [selected]);

  const animateClose = useCallback(() => {
    setClosing(true);
    setTimeout(onClose, 150);
  }, [onClose]);

  // Global escape key listener (catches Escape even when input is focused)
  useEffect(() => {
    if (!open || closing) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        animateClose();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, closing, animateClose]);

  const go = useCallback(
    (hit: SearchHit) => {
      const url = hit.hash ? `/${hit.path}#${hit.hash}` : `/${hit.path}`;
      navigate(url);
      // After navigation, scroll to the hash target
      if (hit.hash) {
        setTimeout(() => {
          const el = document.getElementById(hit.hash);
          if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
        }, 150);
      }
      animateClose();
    },
    [navigate, animateClose],
  );

  // Keyboard navigation
  const onKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          setSelected((s) => Math.min(s + 1, results.length - 1));
          break;
        case "ArrowUp":
          e.preventDefault();
          setSelected((s) => Math.max(s - 1, 0));
          break;
        case "Enter":
          e.preventDefault();
          if (results[selected]) go(results[selected]);
          break;
        case "Escape":
          e.preventDefault();
          animateClose();
          break;
      }
    },
    [results, selected, go, animateClose],
  );

  if (!open) return null;

  const isDark = theme === "dark";

  // Highlight matching terms in snippet
  const highlightSnippet = (snippet: string, q: string) => {
    if (!q.trim()) return snippet;
    const terms = q
      .trim()
      .split(/\s+/)
      .filter((t) => t.length > 1);
    if (terms.length === 0) return snippet;

    const pattern = new RegExp(`(${terms.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|")})`, "gi");
    const parts = snippet.split(pattern);

    return parts.map((part, i) =>
      pattern.test(part) ? (
        <span key={i} style={{ color: t.accent, fontWeight: 600 }}>
          {part}
        </span>
      ) : (
        part
      ),
    );
  };

  const isMac = typeof navigator !== "undefined" && /Mac|iPhone|iPad/.test(navigator.userAgent);

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center"
      style={{
        paddingTop: "min(20vh, 160px)",
        background: isDark ? "rgba(0, 0, 0, 0.6)" : "rgba(0, 0, 0, 0.3)",
        backdropFilter: "blur(8px)",
        WebkitBackdropFilter: "blur(8px)",
        opacity: closing ? 0 : 1,
        transition: "opacity 0.15s ease-out",
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) animateClose();
      }}
    >
      <div
        ref={modalRef}
        style={{
          width: "min(560px, calc(100vw - 32px))",
          maxHeight: "min(480px, 60vh)",
          background: isDark ? "rgba(39, 40, 34, 0.98)" : "rgba(255, 255, 255, 0.98)",
          border: `1px solid ${isDark ? "#49483e" : "#d8d8d0"}`,
          borderRadius: 8,
          boxShadow: isDark
            ? "0 24px 80px rgba(0,0,0,0.7), 0 0 0 1px rgba(102,217,239,0.08)"
            : "0 24px 80px rgba(0,0,0,0.18), 0 0 0 1px rgba(47,158,184,0.08)",
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
          transform: closing ? "scale(0.97) translateY(-4px)" : "scale(1) translateY(0)",
          opacity: closing ? 0 : 1,
          transition: "transform 0.15s ease-out, opacity 0.15s ease-out",
          animation: closing ? "none" : "searchModalIn 0.2s ease-out",
        }}
        onKeyDown={onKeyDown}
      >
        {/* Search input row */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            padding: "12px 16px",
            borderBottom: `1px solid ${t.border}`,
          }}
        >
          {/* Prompt chevron */}
          <span
            style={{
              fontFamily: "'Fira Code', monospace",
              fontSize: 14,
              color: t.accent,
              flexShrink: 0,
              fontWeight: 600,
            }}
          >
            {"\u203a"}
          </span>

          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={onKeyDown}
            onBlur={(e) => {
              // Browser consumes Escape by blurring the input — if focus
              // moved outside the modal (or to nothing), treat it as close.
              const next = e.relatedTarget as Node | null;
              if (!next || !modalRef.current?.contains(next)) {
                animateClose();
              }
            }}
            placeholder="search docs..."
            spellCheck={false}
            autoComplete="off"
            style={{
              flex: 1,
              background: "transparent",
              border: "none",
              outline: "none",
              fontFamily: "'Fira Code', monospace",
              fontSize: 14,
              color: t.text,
              caretColor: t.accent,
              letterSpacing: "-0.01em",
            }}
          />

          {/* Escape hint */}
          <kbd
            style={{
              fontFamily: "'Fira Code', monospace",
              fontSize: 10,
              color: t.textDim,
              background: isDark ? "#3e3d3222" : "#e8e8e322",
              border: `1px solid ${t.border}`,
              borderRadius: 4,
              padding: "2px 6px",
              flexShrink: 0,
            }}
          >
            esc
          </kbd>
        </div>

        {/* Results */}
        <div
          ref={listRef}
          className="custom-scroll"
          style={{
            flex: 1,
            overflowY: "auto",
            padding: results.length > 0 ? "6px" : 0,
          }}
        >
          {query.trim() && results.length === 0 && (
            <div
              style={{
                padding: "32px 16px",
                textAlign: "center",
                fontFamily: "'Fira Code', monospace",
                fontSize: 13,
                color: t.textDim,
              }}
            >
              no results for "{query}"
            </div>
          )}

          {results.map((hit, i) => {
            const isSelected = i === selected;
            const isGuide = hit.type === "guide";
            const badgeColor = isGuide ? t.secondary : t.accent;
            const resultKey = `${hit.path}#${hit.hash}`;

            return (
              <button
                key={resultKey}
                onClick={() => go(hit)}
                onMouseEnter={() => setSelected(i)}
                style={{
                  display: "block",
                  width: "100%",
                  textAlign: "left",
                  padding: "8px 12px",
                  borderRadius: 6,
                  border: "none",
                  cursor: "pointer",
                  fontFamily: "'Fira Code', monospace",
                  background: isSelected
                    ? isDark
                      ? "rgba(102, 217, 239, 0.08)"
                      : "rgba(47, 158, 184, 0.08)"
                    : "transparent",
                  transition: "background 0.1s ease",
                }}
              >
                {/* Title row */}
                <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
                  {/* Type badge */}
                  <span
                    style={{
                      fontSize: 9,
                      fontWeight: 600,
                      textTransform: "uppercase",
                      letterSpacing: "0.08em",
                      color: badgeColor,
                      background: `${badgeColor}14`,
                      border: `1px solid ${badgeColor}33`,
                      borderRadius: 3,
                      padding: "1px 5px",
                      flexShrink: 0,
                    }}
                  >
                    {hit.type}
                  </span>

                  {/* Page title + section breadcrumb */}
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: 500,
                      color: isSelected ? t.text : t.textMuted,
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                      transition: "color 0.1s ease",
                      minWidth: 0,
                    }}
                  >
                    {hit.pageTitle}
                    {hit.sectionTitle && (
                      <>
                        <span style={{ color: t.textDim, margin: "0 5px", fontWeight: 400 }}>{"\u203a"}</span>
                        <span style={{ fontWeight: 400 }}>{hit.sectionTitle}</span>
                      </>
                    )}
                  </span>

                  {/* Enter hint on selected */}
                  {isSelected && (
                    <span
                      style={{
                        marginLeft: "auto",
                        fontSize: 10,
                        color: t.textDim,
                        flexShrink: 0,
                      }}
                    >
                      {"\u23ce"}
                    </span>
                  )}
                </div>

                {/* Snippet */}
                {hit.snippet && (
                  <div
                    style={{
                      fontSize: 11,
                      color: t.textDim,
                      marginTop: 3,
                      lineHeight: 1.5,
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {highlightSnippet(hit.snippet, query)}
                  </div>
                )}
              </button>
            );
          })}
        </div>

        {/* Footer */}
        {!query.trim() && (
          <div
            style={{
              padding: "10px 16px",
              borderTop: `1px solid ${t.border}`,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 16,
              fontFamily: "'Fira Code', monospace",
              fontSize: 10,
              color: t.textDim,
            }}
          >
            <span style={{ display: "flex", alignItems: "center", gap: 4 }}>
              <kbd
                style={{
                  background: isDark ? "#3e3d3244" : "#e8e8e344",
                  border: `1px solid ${t.border}`,
                  borderRadius: 3,
                  padding: "1px 4px",
                  fontSize: 9,
                }}
              >
                {"\u2191\u2193"}
              </kbd>
              navigate
            </span>
            <span style={{ display: "flex", alignItems: "center", gap: 4 }}>
              <kbd
                style={{
                  background: isDark ? "#3e3d3244" : "#e8e8e344",
                  border: `1px solid ${t.border}`,
                  borderRadius: 3,
                  padding: "1px 4px",
                  fontSize: 9,
                }}
              >
                {"\u23ce"}
              </kbd>
              open
            </span>
            <span style={{ display: "flex", alignItems: "center", gap: 4 }}>
              <kbd
                style={{
                  background: isDark ? "#3e3d3244" : "#e8e8e344",
                  border: `1px solid ${t.border}`,
                  borderRadius: 3,
                  padding: "1px 4px",
                  fontSize: 9,
                }}
              >
                {isMac ? "\u2318K" : "^K"}
              </kbd>
              search
            </span>
          </div>
        )}
      </div>

      {/* Modal entrance animation */}
      <style>{`
        @keyframes searchModalIn {
          from {
            opacity: 0;
            transform: scale(0.96) translateY(-8px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
          }
        }
      `}</style>
    </div>
  );
}
