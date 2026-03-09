import { useState, useEffect } from "react";
import { palette, useTheme } from "../lib/theme.ts";

export default function EditorSimulation({
  activeFeature,
  onSetFeature,
  pausedRef,
}: {
  activeFeature: number;
  onSetFeature: (i: number) => void;
  pausedRef: React.RefObject<boolean>;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [showCompletions, setShowCompletions] = useState(false);
  const [showDiagnostic, setShowDiagnostic] = useState(false);
  const [highlightLine, setHighlightLine] = useState<number | null>(null);
  const [showFormatted, setShowFormatted] = useState(false);

  const fmtColor = theme === "dark" ? "#ae81ff" : "#7c5cb8";
  const features = [
    { id: "syntax", label: "syntax highlighting", icon: "\u2726", color: t.accent },
    { id: "completions", label: "completions", icon: "\u00bb", color: t.secondary },
    { id: "diagnostics", label: "diagnostics", icon: "\u26a0", color: t.tertiary },
    { id: "goto", label: "go-to-definition", icon: "\u2192", color: theme === "dark" ? "#e6db74" : "#998a00" },
    { id: "format", label: "auto-format", icon: "\u2261", color: fmtColor },
  ];

  useEffect(() => {
    setShowCompletions(false);
    setShowDiagnostic(false);
    setHighlightLine(null);
    setShowFormatted(false);

    const timer = setTimeout(() => {
      if (activeFeature === 1) setShowCompletions(true);
      if (activeFeature === 2) setShowDiagnostic(true);
      if (activeFeature === 3) setHighlightLine(10);
      if (activeFeature === 4) setShowFormatted(true);
    }, 300);

    return () => clearTimeout(timer);
  }, [activeFeature]);

  const editorLines = [
    { num: 1, tokens: [{ text: "package", color: t.codeKeyword }, { text: " dashboard", color: t.text }] },
    { num: 2, tokens: [{ text: "", color: t.text }] },
    { num: 3, tokens: [{ text: "import", color: t.codeKeyword }, { text: " (", color: t.codePunct }] },
    { num: 4, tokens: [{ text: '  "fmt"', color: t.codeString }] },
    { num: 5, tokens: [{ text: ")", color: t.codePunct }] },
    { num: 6, tokens: [{ text: "", color: t.text }] },
    { num: 7, tokens: [{ text: "templ", color: t.codeKeyword }, { text: " ", color: t.text }, { text: "Dashboard", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: "title ", color: t.text }, { text: "string", color: t.codeKeyword }, { text: ") {", color: t.codePunct }] },
    { num: 8, tokens: [{ text: '  <', color: t.codePunct }, { text: 'div', color: t.codeKeyword }, { text: ' class=', color: t.codePunct }, { text: '"flex-col h-full"', color: t.codeString }, { text: '>', color: t.codePunct }] },
    { num: 9, tokens: [{ text: "    @", color: t.codeDirective }, { text: "Header", color: t.codeFunc }, { text: "(title)", color: t.codePunct }] },
    { num: 10, tokens: [{ text: "    @", color: t.codeDirective }, { text: "Sidebar", color: t.codeFunc }, { text: "()", color: t.codePunct }] },
    { num: 11, tokens: [{ text: "    @", color: t.codeDirective }, { text: "MainContent", color: t.codeFunc }, { text: "()", color: t.codePunct }] },
    { num: 12, tokens: [{ text: "  </", color: t.codePunct }, { text: "div", color: t.codeKeyword }, { text: ">", color: t.codePunct }] },
    { num: 13, tokens: [{ text: "}", color: t.codePunct }] },
    { num: 14, tokens: [{ text: "", color: t.text }] },
    { num: 15, tokens: [{ text: "templ", color: t.codeKeyword }, { text: " ", color: t.text }, { text: "Header", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: "title ", color: t.text }, { text: "string", color: t.codeKeyword }, { text: ") {", color: t.codePunct }] },
    { num: 16, tokens: [{ text: '  <', color: t.codePunct }, { text: 'div', color: t.codeKeyword }, { text: ' class=', color: t.codePunct }, { text: '"border-single p-1"', color: t.codeString }, { text: '>', color: t.codePunct }] },
    { num: 17, tokens: [{ text: '    <', color: t.codePunct }, { text: 'span', color: t.codeKeyword }, { text: ' class=', color: t.codePunct }, { text: '"font-bold text-cyan"', color: t.codeString }, { text: '>', color: t.codePunct }] },
    { num: 18, tokens: [{ text: "      {", color: t.codePunct }, { text: "fmt", color: t.text }, { text: ".", color: t.codePunct }, { text: "Sprintf", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: '"%s"', color: t.codeString }, { text: ", title)", color: t.codePunct }, { text: "}", color: t.codePunct }] },
    { num: 19, tokens: [{ text: "    </", color: t.codePunct }, { text: "span", color: t.codeKeyword }, { text: ">", color: t.codePunct }] },
    { num: 20, tokens: [{ text: "  </", color: t.codePunct }, { text: "div", color: t.codeKeyword }, { text: ">", color: t.codePunct }] },
    { num: 21, tokens: [{ text: "}", color: t.codePunct }] },
  ];

  // Lines with bad indentation (for the format demo)
  // Only certain lines differ — we track which line nums changed
  const fmtChangedLines = new Set([9, 10, 11, 17, 18, 19]);
  const messyLines: typeof editorLines = editorLines.map((line) => {
    if (line.num === 9) return { num: 9, tokens: [{ text: "  @", color: t.codeDirective }, { text: "Header", color: t.codeFunc }, { text: "(title)", color: t.codePunct }] };
    if (line.num === 10) return { num: 10, tokens: [{ text: "      @", color: t.codeDirective }, { text: "Sidebar", color: t.codeFunc }, { text: "()", color: t.codePunct }] };
    if (line.num === 11) return { num: 11, tokens: [{ text: "   @", color: t.codeDirective }, { text: "MainContent", color: t.codeFunc }, { text: "()", color: t.codePunct }] };
    if (line.num === 17) return { num: 17, tokens: [{ text: "  <", color: t.codePunct }, { text: "span", color: t.codeKeyword }, { text: " class=", color: t.codePunct }, { text: '"font-bold text-cyan"', color: t.codeString }, { text: ">", color: t.codePunct }] };
    if (line.num === 18) return { num: 18, tokens: [{ text: "        {", color: t.codePunct }, { text: "fmt", color: t.text }, { text: ".", color: t.codePunct }, { text: "Sprintf", color: t.codeFunc }, { text: "(", color: t.codePunct }, { text: '"%s"', color: t.codeString }, { text: ", title)", color: t.codePunct }, { text: "}", color: t.codePunct }] };
    if (line.num === 19) return { num: 19, tokens: [{ text: "      </", color: t.codePunct }, { text: "span", color: t.codeKeyword }, { text: ">", color: t.codePunct }] };
    return line;
  });

  const completionItems = [
    { label: "Sidebar", detail: "() *Element" },
    { label: "SearchBar", detail: "(query string) *Element" },
    { label: "StatusLine", detail: "() *Element" },
  ];

  return (
    <div
      className="rounded-lg overflow-hidden"
      style={{
        background: t.bgCode,
        border: `1px solid ${t.border}`,
        boxShadow: theme === "dark"
          ? "0 4px 24px rgba(0,0,0,0.4)"
          : "0 2px 12px rgba(0,0,0,0.08)",
      }}
      onMouseEnter={() => { pausedRef.current = true; }}
      onMouseLeave={() => { pausedRef.current = false; }}
    >
      {/* Editor title bar */}
      <div
        className="flex items-center justify-between px-4 py-2"
        style={{ borderBottom: `1px solid ${t.border}` }}
      >
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#ff5f57" }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#febc2e" }} />
            <div className="w-2.5 h-2.5 rounded-full" style={{ background: "#28c840" }} />
          </div>
          <span className="font-['Fira_Code',monospace] text-[10px] ml-2" style={{ color: t.textDim }}>
            dashboard.gsx
          </span>
        </div>
        <div className="flex items-center gap-1.5">
          <span
            className="font-['Fira_Code',monospace] text-[9px] px-1.5 py-0.5 rounded"
            style={{
              color: t.secondary,
              background: `${t.secondary}12`,
              border: `1px solid ${t.secondary}25`,
            }}
          >
            LSP
          </span>
          <span
            className="font-['Fira_Code',monospace] text-[9px] px-1.5 py-0.5 rounded"
            style={{
              color: t.accent,
              background: `${t.accent}12`,
              border: `1px solid ${t.accent}25`,
            }}
          >
            tree-sitter
          </span>
        </div>
      </div>

      {/* Feature selector tabs */}
      <div
        className="flex items-center gap-1 px-3 py-1.5 overflow-x-auto custom-scroll"
        style={{
          borderBottom: `1px solid ${t.border}`,
          background: theme === "dark" ? "#1e1f1a" : "#eeeee8",
        }}
      >
        {features.map((f, i) => (
          <button
            key={f.id}
            onClick={() => onSetFeature(i)}
            className="font-['Fira_Code',monospace] text-[10px] sm:text-[11px] px-2.5 py-1 rounded transition-all duration-200 flex items-center gap-1.5 whitespace-nowrap shrink-0"
            style={{
              color: activeFeature === i ? f.color : t.textDim,
              background: activeFeature === i ? `${f.color}10` : "transparent",
              border: `1px solid ${activeFeature === i ? `${f.color}30` : "transparent"}`,
              cursor: "pointer",
            }}
          >
            <span style={{ fontSize: "9px" }}>{f.icon}</span>
            {f.label}
          </button>
        ))}
      </div>

      {/* Editor body */}
      <div className="relative px-0 py-3 font-['Fira_Code',monospace] text-[11px] sm:text-[12px] leading-[1.8] overflow-x-auto custom-scroll">
        {(activeFeature === 4 && !showFormatted ? messyLines : editorLines).map((line) => {
          const isGotoTarget = activeFeature === 3 && highlightLine === line.num;
          const hasDiagnostic = activeFeature === 2 && showDiagnostic && line.num === 10;
          const isFmtChanged = activeFeature === 4 && showFormatted && fmtChangedLines.has(line.num);

          return (
            <div
              key={line.num}
              className="flex transition-all duration-300"
              style={{
                background: isGotoTarget
                  ? `${features[3].color}10`
                  : hasDiagnostic
                    ? `${t.tertiary}08`
                    : isFmtChanged
                      ? `${fmtColor}08`
                      : "transparent",
                borderLeft: isGotoTarget
                  ? `2px solid ${features[3].color}`
                  : isFmtChanged
                    ? `2px solid ${fmtColor}`
                    : "2px solid transparent",
              }}
            >
              <span
                className="inline-block w-8 sm:w-10 text-right pr-3 sm:pr-4 select-none shrink-0"
                style={{ color: t.textDim, opacity: 0.5 }}
              >
                {line.num}
              </span>
              <span className="whitespace-pre">
                {line.tokens.map((tok, j) => {
                  const isSyntaxHighlighted = activeFeature === 0
                    && tok.text.trim().length > 0
                    && tok.color !== t.text
                    && tok.color !== t.codePunct;
                  return (
                    <span
                      key={j}
                      className={isSyntaxHighlighted ? "syntax-active-token" : ""}
                      style={{
                        color: tok.color,
                        animationDelay: isSyntaxHighlighted ? `${j * 120}ms` : undefined,
                      }}
                    >{tok.text}</span>
                  );
                })}
              </span>
              {hasDiagnostic && (
                <span className="ml-4 text-[10px] flex items-center gap-1.5" style={{ color: t.tertiary }}>
                  <span className="opacity-80">undefined: Sidebar</span>
                </span>
              )}
            </div>
          );
        })}

        {/* Completions popup */}
        {activeFeature === 1 && showCompletions && (
          <div
            className="absolute rounded-md overflow-hidden"
            style={{
              top: "calc(1.8em * 10 + 12px)",
              left: "calc(10px + 8ch)",
              background: theme === "dark" ? "#3e3d32" : "#ffffff",
              border: `1px solid ${t.border}`,
              boxShadow: theme === "dark"
                ? "0 4px 16px rgba(0,0,0,0.5)"
                : "0 4px 16px rgba(0,0,0,0.12)",
              zIndex: 10,
              animation: "fadeInUp 0.2s ease-out forwards",
              minWidth: "220px",
            }}
          >
            {completionItems.map((item, i) => (
              <div
                key={i}
                className="flex items-center gap-2 px-2.5 py-1.5 text-[11px]"
                style={{
                  background: i === 0 ? `${t.accent}15` : "transparent",
                  borderLeft: i === 0 ? `2px solid ${t.accent}` : "2px solid transparent",
                }}
              >
                <span
                  className="px-1 py-0.5 rounded text-[9px]"
                  style={{ background: `${t.secondary}18`, color: t.secondary }}
                >
                  C
                </span>
                <span style={{ color: i === 0 ? t.accent : t.text }}>{item.label}</span>
                <span className="ml-auto" style={{ color: t.textDim, fontSize: "10px" }}>
                  {item.detail}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Status bar */}
      <div
        className="flex items-center justify-between px-4 py-1.5 font-['Fira_Code',monospace] text-[10px]"
        style={{
          borderTop: `1px solid ${t.border}`,
          background: theme === "dark" ? "#1e1f1a" : "#eeeee8",
        }}
      >
        <div className="flex items-center gap-3">
          <span style={{ color: t.textDim }}>Ln 10, Col 5</span>
          <span style={{ color: t.textDim }}>GSX</span>
        </div>
        <div className="flex items-center gap-2">
          {activeFeature === 2 && showDiagnostic && (
            <span className="flex items-center gap-1" style={{ color: t.tertiary }}>
              <svg width="10" height="10" viewBox="0 0 10 10" fill="currentColor">
                <circle cx="5" cy="5" r="4" />
              </svg>
              1 error
            </span>
          )}
          {activeFeature === 4 && showFormatted && (
            <span className="flex items-center gap-1" style={{ color: fmtColor }}>
              6 lines formatted
            </span>
          )}
          <span style={{ color: t.secondary }}>tui lsp</span>
        </div>
      </div>
    </div>
  );
}
