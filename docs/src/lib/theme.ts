import { createContext, useContext } from "react";

export type Theme = "dark" | "light";

export const palette = {
  dark: {
    bg: "#050510",
    bgSecondary: "#0a0a1a",
    bgTertiary: "#0f0f25",
    bgCard: "#0a0a20",
    bgCode: "#06061a",
    text: "#c0c8e0",
    textMuted: "#606888",
    textDim: "#3a4060",
    heading: "#e0e8ff",
    accent: "#00ffff",
    accentDim: "#00aaaa",
    accentGlow: "0 0 7px #00ffff, 0 0 20px #00ffff44, 0 0 40px #00ffff22",
    accentGlowSubtle: "0 0 5px #00ffff88, 0 0 15px #00ffff22",
    secondary: "#39ff14",
    secondaryDim: "#22aa0d",
    secondaryGlow: "0 0 7px #39ff14, 0 0 20px #39ff1444",
    tertiary: "#ff00ff",
    tertiaryDim: "#aa00aa",
    tertiaryGlow: "0 0 7px #ff00ff, 0 0 20px #ff00ff44",
    border: "#1a1a3a",
    borderGlow: "#00ffff33",
    codeKeyword: "#ff00ff",
    codeString: "#39ff14",
    codeComment: "#3a4060",
    codeFunc: "#00ffff",
    codePunct: "#606888",
    codeDirective: "#ff00ff",
  },
  light: {
    bg: "#f0f4f8",
    bgSecondary: "#e8ecf2",
    bgTertiary: "#dfe5ed",
    bgCard: "#ffffff",
    bgCode: "#f8fafc",
    text: "#2a2a4e",
    textMuted: "#5a5a7e",
    textDim: "#8a8aaa",
    heading: "#0a0a2e",
    accent: "#0088aa",
    accentDim: "#006688",
    accentGlow: "none",
    accentGlowSubtle: "none",
    secondary: "#1a8a0a",
    secondaryDim: "#15700a",
    secondaryGlow: "none",
    tertiary: "#aa00aa",
    tertiaryDim: "#880088",
    tertiaryGlow: "none",
    border: "#d0d8e4",
    borderGlow: "#0088aa22",
    codeKeyword: "#aa00aa",
    codeString: "#1a8a0a",
    codeComment: "#8a8aaa",
    codeFunc: "#0088aa",
    codePunct: "#5a5a7e",
    codeDirective: "#aa00aa",
  },
};

export const ThemeContext = createContext<{
  theme: Theme;
  setTheme: (t: Theme) => void;
}>({
  theme: "dark",
  setTheme: () => {},
});

export function useTheme() {
  return useContext(ThemeContext);
}
