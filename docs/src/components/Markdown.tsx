import { useState, useEffect, useMemo, type ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useTheme, palette } from "../lib/theme.ts";
import { getHighlighter, highlight } from "../lib/highlighter.ts";

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

export interface HeadingEntry {
  level: number;
  text: string;
  id: string;
  line: number;
}

/**
 * Pre-compute all heading IDs from raw markdown. Both the Markdown renderer
 * and the TableOfContents must use this same function so IDs always match.
 */
export function extractHeadingIds(markdown: string): HeadingEntry[] {
  const entries: HeadingEntry[] = [];
  const seen = new Map<string, number>();
  let inCodeBlock = false;
  const lines = markdown.split("\n");

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (line.trimStart().startsWith("```")) {
      inCodeBlock = !inCodeBlock;
      continue;
    }
    if (inCodeBlock) continue;

    const m = line.match(/^(#{2,3})\s+(.+)/);
    if (m) {
      const level = m[1].length;
      const text = m[2].replace(/[*_`]/g, "").trim();
      const base = slugify(text);
      const count = (seen.get(base) ?? 0) + 1;
      seen.set(base, count);
      const id = count === 1 ? base : `${base}-${count}`;
      entries.push({ level, text, id, line: i + 1 });
    }
  }

  return entries;
}

function childrenToText(children: ReactNode): string {
  if (typeof children === "string") return children;
  if (typeof children === "number") return String(children);
  if (Array.isArray(children)) return children.map(childrenToText).join("");
  if (children && typeof children === "object" && "props" in children) {
    return childrenToText((children as any).props.children);
  }
  return "";
}

function Markdown({ content }: { content: string }) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [ready, setReady] = useState(false);

  useEffect(() => {
    getHighlighter().then(() => setReady(true));
  }, []);

  const idByLine = useMemo(() => {
    const map = new Map<number, string>();
    for (const entry of extractHeadingIds(content)) {
      map.set(entry.line, entry.id);
    }
    return map;
  }, [content]);

  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        h1: ({ children }) => (
          <h1
            className="text-2xl sm:text-3xl font-bold tracking-tight mb-6 sm:mb-8"
            style={{ color: t.heading }}
          >
            {children}
          </h1>
        ),
        h2: ({ children, node }) => {
          const line = (node as any)?.position?.start?.line;
          const id = line ? (idByLine.get(line) ?? "") : "";
          return (
            <h2
              id={id}
              className="group text-base sm:text-lg font-semibold mb-3 mt-10 sm:mt-12 scroll-mt-16"
              style={{ color: t.heading }}
            >
              {children}
              {id && (
                <a
                  href={`#${id}`}
                  onClick={(e) => {
                    e.preventDefault();
                    history.replaceState(null, "", `#${id}`);
                    const url = `${window.location.origin}${window.location.pathname}#${id}`;
                    navigator.clipboard.writeText(url);
                  }}
                  className="ml-2 transition-colors duration-150 no-underline"
                  style={{ color: t.textDim, textDecoration: "none", opacity: 0.4 }}
                  onMouseEnter={(e) => { e.currentTarget.style.opacity = "1"; e.currentTarget.style.color = t.accent; }}
                  onMouseLeave={(e) => { e.currentTarget.style.opacity = "0.4"; e.currentTarget.style.color = t.textDim; }}
                  aria-label="Copy link to section"
                >
                  #
                </a>
              )}
            </h2>
          );
        },
        h3: ({ children, node }) => {
          const line = (node as any)?.position?.start?.line;
          const id = line ? (idByLine.get(line) ?? "") : "";
          return (
            <h3
              id={id}
              className="group text-[14px] sm:text-[15px] font-semibold mb-2 mt-6 scroll-mt-16"
              style={{ color: t.heading }}
            >
              {children}
              {id && (
                <a
                  href={`#${id}`}
                  onClick={(e) => {
                    e.preventDefault();
                    history.replaceState(null, "", `#${id}`);
                    const url = `${window.location.origin}${window.location.pathname}#${id}`;
                    navigator.clipboard.writeText(url);
                  }}
                  className="ml-2 transition-colors duration-150 no-underline"
                  style={{ color: t.textDim, textDecoration: "none", opacity: 0.4 }}
                  onMouseEnter={(e) => { e.currentTarget.style.opacity = "1"; e.currentTarget.style.color = t.accent; }}
                  onMouseLeave={(e) => { e.currentTarget.style.opacity = "0.4"; e.currentTarget.style.color = t.textDim; }}
                  aria-label="Copy link to section"
                >
                  #
                </a>
              )}
            </h3>
          );
        },
        p: ({ children }) => (
          <p
            className="text-[13px] sm:text-[14px] leading-relaxed mb-4"
            style={{ color: t.textMuted }}
          >
            {children}
          </p>
        ),
        a: ({ href, children }) => (
          <a
            href={href}
            className="transition-colors duration-200"
            style={{ color: t.accent }}
            onMouseEnter={(e) => (e.currentTarget.style.color = t.accentDim)}
            onMouseLeave={(e) => (e.currentTarget.style.color = t.accent)}
            target={href?.startsWith("http") ? "_blank" : undefined}
            rel={href?.startsWith("http") ? "noopener noreferrer" : undefined}
          >
            {children}
          </a>
        ),
        ul: ({ children }) => (
          <ul
            className="text-[13px] sm:text-[14px] leading-relaxed mb-4 ml-4 list-disc"
            style={{ color: t.textMuted }}
          >
            {children}
          </ul>
        ),
        ol: ({ children }) => (
          <ol
            className="text-[13px] sm:text-[14px] leading-relaxed mb-4 ml-4 list-decimal"
            style={{ color: t.textMuted }}
          >
            {children}
          </ol>
        ),
        li: ({ children }) => <li className="mb-1">{children}</li>,
        blockquote: ({ children }) => (
          <blockquote
            className="border-l-2 pl-4 my-4"
            style={{ borderColor: t.accent, color: t.textMuted }}
          >
            {children}
          </blockquote>
        ),
        table: ({ children }) => (
          <ResponsiveTable theme={theme}>{children}</ResponsiveTable>
        ),
        thead: ({ children }) => (
          <thead style={{ borderBottom: `2px solid ${t.border}` }}>
            {children}
          </thead>
        ),
        th: ({ children }) => (
          <th
            className="font-['Fira_Code',monospace] text-left px-3 py-2 font-semibold"
            style={{ color: t.heading }}
          >
            {children}
          </th>
        ),
        tr: ({ children }) => {
          const [hovered, setHovered] = useState(false);
          return (
            <tr
              style={{
                borderBottom: `1px solid ${t.border}`,
                background: hovered ? t.bgTertiary : "transparent",
              }}
              onMouseEnter={() => setHovered(true)}
              onMouseLeave={() => setHovered(false)}
            >
              {children}
            </tr>
          );
        },
        td: ({ children }) => (
          <td className="px-3 py-2" style={{ color: t.textMuted }}>
            {children}
          </td>
        ),
        code: ({ className, children, ...props }) => {
          const match = /language-(\w+)/.exec(className ?? "");
          const lang = match ? match[1] : null;
          const codeStr = String(children).replace(/\n$/, "");
          const isBlock = codeStr.includes("\n");

          if (!lang && !isBlock) {
            return (
              <code
                className="font-['Fira_Code',monospace] text-[12px] px-1.5 py-0.5 rounded"
                style={{
                  background:
                    theme === "dark" ? "#66d9ef0d" : "#2f9eb80d",
                  color: t.accent,
                  border: `1px solid ${theme === "dark" ? "#66d9ef22" : "#2f9eb822"}`,
                }}
                {...props}
              >
                {children}
              </code>
            );
          }

          // Plain code block (no language) — render as a styled diagram/pre block
          if (!lang && isBlock) {
            return (
              <DiagramBlock content={codeStr} theme={theme} />
            );
          }

          const html = ready ? highlight(codeStr, lang!, theme) : "";

          return (
            <CodeBlockWrapper lang={lang!} title={lang ?? undefined} code={codeStr}>
              {html ? (
                <div
                  className="shiki-container [&_pre]:!bg-transparent [&_pre]:!m-0 [&_pre]:!p-0 [&_code]:!text-[12px] [&_code]:!sm:text-[13px] [&_code]:!leading-[1.7] [&_code]:font-['Fira_Code',monospace]"
                  dangerouslySetInnerHTML={{ __html: html }}
                />
              ) : (
                <pre className="font-['Fira_Code',monospace] text-[12px] sm:text-[13px] leading-[1.7]">
                  <code style={{ color: t.text }}>{codeStr}</code>
                </pre>
              )}
            </CodeBlockWrapper>
          );
        },
        hr: () => (
          <div className="my-8">
            <div
              className="h-px"
              style={{
                background:
                  theme === "dark"
                    ? "linear-gradient(to right, transparent, #66d9ef18, #f9267218, transparent)"
                    : `linear-gradient(to right, transparent, ${t.border}, transparent)`,
              }}
            />
          </div>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}

/* ─── Responsive Table ─── */

function ResponsiveTable({ children, theme }: { children: ReactNode; theme: string }) {
  const t = palette[theme as "dark" | "light"];

  // Extract header labels and body rows from the React children tree
  const headers: string[] = [];
  const bodyRows: ReactNode[][] = [];

  const childArray = Array.isArray(children) ? children : [children];
  for (const child of childArray) {
    if (!child || typeof child !== "object" || !("props" in child)) continue;
    const el = child as any;
    const tag = el.type?.name ?? el.type ?? el.props?.node?.tagName;

    if (tag === "thead" || el.props?.node?.tagName === "thead") {
      // Walk thead > tr > th to get header text
      const walkTh = (node: any) => {
        if (!node) return;
        const kids = Array.isArray(node.props?.children) ? node.props.children : [node.props?.children];
        for (const k of kids) {
          if (!k || typeof k !== "object") continue;
          const kTag = (k as any).type?.name ?? (k as any).type ?? (k as any).props?.node?.tagName;
          if (kTag === "th" || (k as any).props?.node?.tagName === "th") {
            headers.push(childrenToText((k as any).props?.children));
          } else {
            walkTh(k);
          }
        }
      };
      walkTh(el);
    }

    if (tag === "tbody" || el.props?.node?.tagName === "tbody") {
      const tbodyKids = Array.isArray(el.props?.children) ? el.props.children : [el.props?.children];
      for (const tr of tbodyKids) {
        if (!tr || typeof tr !== "object") continue;
        const trKids = Array.isArray((tr as any).props?.children) ? (tr as any).props.children : [(tr as any).props?.children];
        const cells: ReactNode[] = [];
        for (const td of trKids) {
          if (td && typeof td === "object" && "props" in td) {
            cells.push((td as any).props?.children);
          }
        }
        if (cells.length > 0) bodyRows.push(cells);
      }
    }
  }

  return (
    <>
      {/* Desktop: normal table */}
      <div className="hidden sm:block overflow-x-auto mb-4">
        <table
          className="w-full text-[12px] sm:text-[13px]"
          style={{ borderCollapse: "collapse" }}
        >
          {children}
        </table>
      </div>

      {/* Mobile: stacked cards */}
      <div className="sm:hidden mb-4 flex flex-col gap-2">
        {bodyRows.map((cells, rowIdx) => (
          <div
            key={rowIdx}
            className="rounded-lg px-3 py-2.5 text-[12px]"
            style={{
              background: t.bgSecondary,
              border: `1px solid ${t.border}`,
            }}
          >
            {cells.map((cell, cellIdx) => (
              <div key={cellIdx} className={cellIdx > 0 ? "mt-1.5" : ""}>
                {headers[cellIdx] && (
                  <div
                    className="font-['Fira_Code',monospace] text-[10px] uppercase tracking-wider mb-0.5"
                    style={{ color: t.textDim }}
                  >
                    {headers[cellIdx]}
                  </div>
                )}
                <div style={{ color: cellIdx === 0 ? t.heading : t.textMuted }}>
                  {cell}
                </div>
              </div>
            ))}
          </div>
        ))}
      </div>
    </>
  );
}

/* ─── Diagram Block ─── */

function DiagramBlock({ content, theme }: { content: string; theme: string }) {
  const t = palette[theme as "dark" | "light"];

  // Parse the pipeline: split on arrow sequences to extract stages
  const hasArrows = /[─═→>]{3,}/.test(content);

  if (hasArrows) {
    // Split into stages by arrow-like sequences
    const parts = content
      .replace(/\n/g, " ")
      .split(/\s*[─═]{2,}>?\s*/)
      .map((s) => s.trim())
      .filter(Boolean);

    if (parts.length >= 2) {
      return (
        <div className="mb-4 flex flex-col sm:flex-row items-stretch sm:items-center gap-2 sm:gap-0">
          {parts.map((part, i) => (
            <div key={i} className="flex flex-col sm:flex-row items-stretch sm:items-center">
              <div
                className="rounded-lg px-3 py-2 text-center font-['Fira_Code',monospace] text-[11px] sm:text-[12px] leading-snug"
                style={{
                  background: t.bgCode,
                  border: `1px solid ${t.border}`,
                  color: i === 0 ? t.accent : i === parts.length - 1 ? t.secondary : t.text,
                  whiteSpace: "pre-line",
                }}
              >
                {part.split(/\s{2,}/).map((line, j) => (
                  <div key={j}>{line}</div>
                ))}
              </div>
              {i < parts.length - 1 && (
                <>
                  <div
                    className="hidden sm:flex items-center px-2"
                    style={{ color: t.textDim }}
                  >
                    <svg width="20" height="12" viewBox="0 0 20 12" fill="none">
                      <path d="M0 6h16M12 1l5 5-5 5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                  </div>
                  <div
                    className="sm:hidden flex justify-center py-0.5"
                    style={{ color: t.textDim }}
                  >
                    <svg width="12" height="16" viewBox="0 0 12 16" fill="none">
                      <path d="M6 0v12M1 8l5 5 5-5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                  </div>
                </>
              )}
            </div>
          ))}
        </div>
      );
    }
  }

  // Fallback: just a styled pre block
  return (
    <div
      className="rounded-lg overflow-hidden mb-4"
      style={{
        background: t.bgCode,
        border: `1px solid ${t.border}`,
      }}
    >
      <div className="px-4 py-3.5 overflow-x-auto custom-scroll">
        <pre className="font-['Fira_Code',monospace] text-[12px] sm:text-[13px] leading-[1.7]">
          <code style={{ color: t.text }}>{content}</code>
        </pre>
      </div>
    </div>
  );
}

function CopyButton({ text }: { text: string }) {
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
      className="p-1 rounded transition-all duration-200"
      style={{
        color: copied ? t.secondary : t.textDim,
        background: "transparent",
        border: "none",
        cursor: "pointer",
      }}
      onMouseEnter={(e) => {
        if (!copied) e.currentTarget.style.color = t.accent;
      }}
      onMouseLeave={(e) => {
        if (!copied) e.currentTarget.style.color = t.textDim;
      }}
      title={copied ? "Copied!" : "Copy code"}
    >
      {copied ? (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      ) : (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      )}
    </button>
  );
}

function CodeBlockWrapper({
  children,
  lang,
  code,
}: {
  children: ReactNode;
  lang: string;
  title?: string;
  code?: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];

  return (
    <div
      className="rounded-lg overflow-hidden mb-4"
      style={{
        background: t.bgCode,
        border: `1px solid ${t.border}`,
        boxShadow:
          theme === "dark"
            ? `0 0 12px ${t.borderGlow}, inset 0 0 20px rgba(0,0,0,0.3)`
            : "0 1px 3px rgba(0,0,0,0.08)",
      }}
    >
      <div
        className="flex items-center justify-between px-4 py-2"
        style={{ borderBottom: `1px solid ${t.border}` }}
      >
        <span
          className="font-['Fira_Code',monospace] text-[10px]"
          style={{ color: t.accentDim }}
        >
          {lang}
        </span>
        {code && <CopyButton text={code} />}
      </div>
      <div className="px-4 py-3.5 overflow-x-auto custom-scroll">{children}</div>
    </div>
  );
}

export default Markdown;
