import { type Theme } from "../lib/theme.ts";

export default function PageBackground({ theme }: { theme: Theme }) {
  const isDark = theme === "dark";
  const lineAlpha = isDark ? "0.025" : "0.018";
  const lineRgb = isDark ? "248,248,242" : "39,40,34";
  const glowColor = isDark
    ? "rgba(166,226,46,0.03)"
    : "rgba(212,37,104,0.02)";

  return (
    <div
      className="absolute inset-0 overflow-hidden pointer-events-none"
      aria-hidden="true"
    >
      {/* Scan lines */}
      <div
        className="absolute inset-0"
        style={{
          top: "-80px",
          bottom: "0",
          backgroundImage: `repeating-linear-gradient(0deg, transparent, transparent 3px, rgba(${lineRgb},${lineAlpha}) 3px, rgba(${lineRgb},${lineAlpha}) 4px)`,
          animation: "scanDrift 12s linear infinite",
          willChange: "transform",
        }}
      />
      {/* Warm radial glow */}
      <div
        className="absolute inset-0"
        style={{
          background: `radial-gradient(ellipse at 25% 40%, ${glowColor} 0%, transparent 55%)`,
        }}
      />
    </div>
  );
}
