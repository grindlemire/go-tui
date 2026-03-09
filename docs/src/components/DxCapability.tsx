import { palette, useTheme } from "../lib/theme.ts";

export default function DxCapability({
  title,
  description,
  color,
  delay,
  active,
  onHover,
  onLeave,
}: {
  title: string;
  description: string;
  color: string;
  delay: number;
  active?: boolean;
  onHover?: () => void;
  onLeave?: () => void;
}) {
  const { theme } = useTheme();
  const t = palette[theme];
  const highlighted = active ?? false;

  return (
    <div
      className="py-2.5 sm:py-3 px-3 sm:px-4 rounded-lg transition-all duration-200 cursor-default"
      style={{
        background: highlighted ? `${color}06` : "transparent",
        borderLeft: `2px solid ${highlighted ? color : "transparent"}`,
        animation: `fadeInUp 0.4s ease-out ${delay}ms both`,
      }}
      onMouseEnter={onHover}
      onMouseLeave={onLeave}
    >
      <div className="flex items-center gap-2 mb-0.5 sm:mb-1">
        <div className="w-1.5 h-1.5 rounded-full shrink-0" style={{ background: color }} />
        <div className="font-['Fira_Code',monospace] text-[12px] sm:text-[13px] font-medium" style={{ color: t.heading }}>{title}</div>
      </div>
      <div className="text-[12px] sm:text-[13px] leading-relaxed pl-3.5" style={{ color: t.textMuted }}>{description}</div>
    </div>
  );
}
