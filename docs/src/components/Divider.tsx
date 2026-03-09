import { palette, useTheme } from "../lib/theme.ts";

export default function Divider() {
  const { theme } = useTheme();
  const t = palette[theme];
  return (
    <div className="max-w-[1100px] mx-auto px-4 sm:px-6 py-6 sm:py-8">
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
  );
}
