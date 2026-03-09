import { useEffect } from "react";
import { useLocation, Outlet } from "react-router-dom";
import { palette, useTheme } from "../lib/theme.ts";
import Nav from "./Nav.tsx";
import Footer from "./Footer.tsx";

function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    history.scrollRestoration = "manual";
    window.scrollTo(0, 0);
  }, [pathname]);
  return null;
}

export { ScrollToTop };

export default function PageShell({ children }: { children: React.ReactNode }) {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div
      className={`${theme === "dark" ? "dark-theme" : "light-theme"} neon-select overflow-x-clip flex flex-col`}
      style={{
        background: t.bg,
        color: t.text,
        minHeight: "100vh",
        fontFamily: "'IBM Plex Sans', sans-serif",
      }}
    >
      <Nav />
      <div className="flex-1">
        {children}
      </div>
      <Footer />
    </div>
  );
}

export function PageLayout() {
  return (
    <PageShell>
      <Outlet />
    </PageShell>
  );
}
