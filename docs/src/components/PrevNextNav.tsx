import { Link } from "react-router-dom";
import { palette, useTheme } from "../lib/theme.ts";

export default function PrevNextNav({
  pages,
  activeIndex,
  basePath,
}: {
  pages: { slug: string; title: string }[];
  activeIndex: number;
  basePath: string;
}) {
  const { theme } = useTheme();
  const t = palette[theme];

  return (
    <div
      className="mt-10 sm:mt-12 pt-6 sm:pt-8 flex justify-between"
      style={{ borderTop: `1px solid ${t.border}` }}
    >
      {activeIndex > 0 ? (
        <Link
          to={`${basePath}/${pages[activeIndex - 1].slug}`}
          className="font-['Fira_Code',monospace] text-xs sm:text-sm transition-colors duration-200"
          style={{
            color: t.textMuted,
            textDecoration: "none",
          }}
          onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
          onMouseLeave={(e) => (e.currentTarget.style.color = t.textMuted)}
        >
          &larr; {pages[activeIndex - 1].title}
        </Link>
      ) : (
        <div />
      )}
      {activeIndex < pages.length - 1 ? (
        <Link
          to={`${basePath}/${pages[activeIndex + 1].slug}`}
          className="font-['Fira_Code',monospace] text-xs sm:text-sm transition-colors duration-200"
          style={{
            color: t.textMuted,
            textDecoration: "none",
          }}
          onMouseEnter={(e) => (e.currentTarget.style.color = t.accent)}
          onMouseLeave={(e) => (e.currentTarget.style.color = t.textMuted)}
        >
          {pages[activeIndex + 1].title} &rarr;
        </Link>
      ) : (
        <div />
      )}
    </div>
  );
}
