import { useEffect } from "react";
import { Link, useParams, useNavigate, useLocation } from "react-router-dom";
import { palette, useTheme } from "../lib/theme.ts";
import { loadGuide } from "../lib/markdown.ts";
import Markdown from "../components/Markdown.tsx";
import TableOfContents from "../components/TableOfContents.tsx";
import PrevNextNav from "../components/PrevNextNav.tsx";
import MobilePicker from "../components/MobilePicker.tsx";
import SidebarLLMButton from "../components/SidebarLLMButton.tsx";
import { RawMarkdownButton } from "../components/CopyButton.tsx";

export default function GuidePage() {
  const { theme } = useTheme();
  const t = palette[theme];
  const { slug } = useParams();
  const navigate = useNavigate();
  const location = useLocation();

  const pages = loadGuide();
  const activeSection = Math.max(0, pages.findIndex((p) => p.slug === slug));

  // Deep link: scroll to hash target on mount / page change
  useEffect(() => {
    const hash = location.hash.replace("#", "");
    if (!hash) return;
    const timer = setTimeout(() => {
      const el = document.getElementById(hash);
      if (el) {
        el.scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }, 100);
    return () => clearTimeout(timer);
  }, [slug, location.hash]);

  return (
    <div className="max-w-[1100px] xl:max-w-[1360px] mx-auto px-4 sm:px-6 pt-10 sm:pt-16 pb-16 sm:pb-24">
      <h1
        className="text-3xl sm:text-5xl font-bold tracking-tight mb-8 sm:mb-12"
        style={{ color: t.heading }}
      >
        Guide
      </h1>

      <div className="flex gap-8 sm:gap-10">
        {/* Desktop Sidebar */}
        <div className="w-48 shrink-0 hidden md:block">
          <div className="sticky top-16">
            <div
              className="font-['Fira_Code',monospace] text-[10px] tracking-[0.15em] uppercase mb-4"
              style={{ color: t.textDim }}
            >
              chapters
            </div>

            {pages.map((page, i) => {
              const active = activeSection === i;
              return (
                <Link
                  key={page.slug}
                  to={`/guide/${page.slug}`}
                  className="block w-full text-left font-['Fira_Code',monospace] text-[12px] py-1.5 px-3 rounded transition-all duration-200"
                  style={{
                    color: active ? t.accent : t.textMuted,
                    background: active
                      ? theme === "dark"
                        ? "#66d9ef0d"
                        : "#2f9eb80d"
                      : "transparent",
                    textDecoration: "none",
                    borderLeft: `2px solid ${active ? t.accent : "transparent"}`,
                  }}
                  onMouseEnter={(e) => {
                    if (!active) e.currentTarget.style.color = t.accent;
                  }}
                  onMouseLeave={(e) => {
                    if (!active) e.currentTarget.style.color = t.textMuted;
                  }}
                >
                  {page.title}
                </Link>
              );
            })}

            <SidebarLLMButton label="copy all as markdown" />
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1 min-w-0">
          <MobilePicker
            pages={pages}
            activeIndex={activeSection}
            onSelect={(i) => navigate(`/guide/${pages[i].slug}`)}
          />

          <RawMarkdownButton body={pages[activeSection].body} />

          <div className="fade-in" key={slug}>
            <Markdown content={pages[activeSection].body} />
          </div>

          <PrevNextNav
            pages={pages}
            activeIndex={activeSection}
            basePath="/guide"
          />
        </div>

        {/* On-page TOC */}
        <TableOfContents content={pages[activeSection].body} key={`toc-${slug}`} />
      </div>
    </div>
  );
}
