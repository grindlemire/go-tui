import { palette, useTheme } from "../lib/theme.ts";

export default function Footer() {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <footer
      className="mt-20"
      style={{ borderTop: `1px solid ${t.border}` }}
    >
      <div className="py-6 flex flex-col items-center justify-center gap-2 font-['Fira_Code',monospace] text-[11px]">
        <div className="flex items-center gap-4">
          <a
            href="https://github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors duration-200"
            style={{ color: t.textDim }}
            onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
            onMouseLeave={(e) => (e.currentTarget.style.color = t.textDim)}
          >
            GitHub
          </a>
          <span style={{ color: t.border }}>&middot;</span>
          <a
            href="https://pkg.go.dev/github.com/grindlemire/go-tui"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-colors duration-200"
            style={{ color: t.textDim }}
            onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
            onMouseLeave={(e) => (e.currentTarget.style.color = t.textDim)}
          >
            pkg.go.dev
          </a>
        </div>
        <span style={{ color: t.textDim }}>
          &copy; {new Date().getFullYear()} Joel Holsteen. All rights reserved.
        </span>
      </div>
    </footer>
  );
}
