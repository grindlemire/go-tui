import { useState, useEffect, type ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useTheme, palette } from "../lib/theme.ts";
import { getHighlighter, highlight } from "../lib/highlighter.ts";

function Markdown({ content }: { content: string }) {
  const { theme } = useTheme();
  const t = palette[theme];
  const [ready, setReady] = useState(false);

  useEffect(() => {
    getHighlighter().then(() => setReady(true));
  }, []);

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
        h2: ({ children }) => (
          <h2
            className="text-base sm:text-lg font-semibold mb-3 mt-10 sm:mt-12"
            style={{ color: t.heading }}
          >
            {children}
          </h2>
        ),
        h3: ({ children }) => (
          <h3
            className="text-[14px] sm:text-[15px] font-semibold mb-2 mt-6"
            style={{ color: t.heading }}
          >
            {children}
          </h3>
        ),
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
          <div className="overflow-x-auto mb-4">
            <table
              className="w-full text-[12px] sm:text-[13px]"
              style={{ borderCollapse: "collapse" }}
            >
              {children}
            </table>
          </div>
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

          if (!lang) {
            return (
              <code
                className="font-['Fira_Code',monospace] text-[12px] px-1.5 py-0.5 rounded"
                style={{
                  background:
                    theme === "dark" ? "#00ffff08" : "#0088aa08",
                  color: t.accent,
                  border: `1px solid ${theme === "dark" ? "#00ffff22" : "#0088aa22"}`,
                }}
                {...props}
              >
                {children}
              </code>
            );
          }

          const html = ready ? highlight(codeStr, lang, theme) : "";

          return (
            <CodeBlockWrapper lang={lang} title={lang}>
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
                    ? "linear-gradient(to right, transparent, #00ffff18, #ff00ff18, transparent)"
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

function CodeBlockWrapper({
  children,
  lang,
  title,
}: {
  children: ReactNode;
  lang: string;
  title?: string;
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
        <div className="flex items-center gap-2">
          <div className="flex gap-1.5">
            <div
              className="w-2.5 h-2.5 rounded-full"
              style={{ background: "#ff5f57" }}
            />
            <div
              className="w-2.5 h-2.5 rounded-full"
              style={{ background: "#febc2e" }}
            />
            <div
              className="w-2.5 h-2.5 rounded-full"
              style={{ background: "#28c840" }}
            />
          </div>
          {title && (
            <span
              className="font-['Fira_Code',monospace] text-[10px] ml-2"
              style={{ color: t.textDim }}
            >
              {title}
            </span>
          )}
        </div>
        <span
          className="font-['Fira_Code',monospace] text-[10px]"
          style={{ color: t.accentDim }}
        >
          {lang}
        </span>
      </div>
      <div className="px-4 py-3.5 overflow-x-auto custom-scroll">{children}</div>
    </div>
  );
}

export default Markdown;
