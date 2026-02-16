import { createContext, useContext } from "react";

export type Theme = "dark" | "light";

export const palette = {
  dark: {
    bg: "#272822",
    bgSecondary: "#2d2e27",
    bgTertiary: "#3e3d32",
    bgCard: "#2d2e27",
    bgCode: "#23241e",
    text: "#f8f8f2",
    textMuted: "#a6a68a",
    textDim: "#75715e",
    heading: "#f8f8f2",
    accent: "#66d9ef",
    accentDim: "#5bb6c8",
    accentGlow: "none",
    accentGlowSubtle: "none",
    secondary: "#a6e22e",
    secondaryDim: "#8aba24",
    secondaryGlow: "none",
    tertiary: "#f92672",
    tertiaryDim: "#d41e62",
    tertiaryGlow: "none",
    border: "#49483e",
    borderGlow: "transparent",
    codeKeyword: "#f92672",
    codeString: "#e6db74",
    codeComment: "#75715e",
    codeFunc: "#a6e22e",
    codePunct: "#a6a68a",
    codeDirective: "#f92672",
  },
  light: {
    bg: "#fafaf8",
    bgSecondary: "#f0f0ec",
    bgTertiary: "#e8e8e3",
    bgCard: "#ffffff",
    bgCode: "#f5f5f1",
    text: "#49483e",
    textMuted: "#75715e",
    textDim: "#a6a68a",
    heading: "#272822",
    accent: "#2f9eb8",
    accentDim: "#267a8f",
    accentGlow: "none",
    accentGlowSubtle: "none",
    secondary: "#638b0c",
    secondaryDim: "#4e6d09",
    secondaryGlow: "none",
    tertiary: "#d42568",
    tertiaryDim: "#b01e57",
    tertiaryGlow: "none",
    border: "#d8d8d0",
    borderGlow: "transparent",
    codeKeyword: "#d42568",
    codeString: "#998a00",
    codeComment: "#a6a68a",
    codeFunc: "#638b0c",
    codePunct: "#75715e",
    codeDirective: "#d42568",
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
