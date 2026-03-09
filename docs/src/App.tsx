import { lazy, Suspense, useState, useEffect, useCallback } from "react";
import { Routes, Route, Navigate } from "react-router-dom";
import { type Theme, ThemeContext } from "./lib/theme.ts";
import { SearchContext } from "./lib/search-context.ts";
import { ScrollToTop } from "./components/PageShell.tsx";
import { PageLayout } from "./components/PageShell.tsx";
import SearchModal from "./components/SearchModal.tsx";

const HomePageExplore = lazy(() => import("./components/HomePageExplore.tsx"));
const GuidePage = lazy(() => import("./pages/GuidePage.tsx"));
const ReferencePage = lazy(() => import("./pages/ReferencePage.tsx"));
const RawGuidePage = lazy(() => import("./pages/RawGuidePage.tsx"));
const GuideRedirect = lazy(() => import("./pages/GuideRedirect.tsx"));
const ReferenceRedirect = lazy(() => import("./pages/ReferenceRedirect.tsx"));

export default function Design2() {
  const [theme, setThemeState] = useState<Theme>(() => {
    const saved = localStorage.getItem("go-tui-theme");
    return saved === "light" || saved === "dark" ? saved : "dark";
  });
  const setTheme = (t: Theme) => {
    localStorage.setItem("go-tui-theme", t);
    setThemeState(t);
  };

  const [searchOpen, setSearchOpen] = useState(false);
  const openSearch = useCallback(() => setSearchOpen(true), []);

  // Global Cmd+K / Ctrl+K shortcut
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(true);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      <SearchContext.Provider value={{ openSearch }}>
        <ScrollToTop />
        <SearchModal open={searchOpen} onClose={() => setSearchOpen(false)} />
        <Suspense>
          <Routes>
            <Route element={<PageLayout />}>
              <Route path="/" element={<HomePageExplore />} />
              <Route path="/guide" element={<GuideRedirect />} />
              <Route path="/guide/:slug" element={<GuidePage />} />
              <Route path="/reference" element={<ReferenceRedirect />} />
              <Route path="/reference/:slug" element={<ReferencePage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
            <Route path="/guide/:slug/raw" element={<RawGuidePage />} />
          </Routes>
        </Suspense>
      </SearchContext.Provider>
    </ThemeContext.Provider>
  );
}
