import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import type { ThemeRegistrationRaw } from "shiki/core";
import { createOnigurumaEngine } from "shiki/engine/oniguruma";
import goLang from "@shikijs/langs/go";
import shellLang from "@shikijs/langs/shellscript";
import gsxGrammar from "../../../editor/vscode/syntaxes/gsx.tmLanguage.json";

const darkTheme: ThemeRegistrationRaw = {
  name: "go-tui-dark",
  type: "dark",
  colors: {
    "editor.background": "#06061a",
    "editor.foreground": "#c0c8e0",
  },
  tokenColors: [
    {
      scope: ["comment", "comment.line", "comment.block"],
      settings: { foreground: "#3a4060" },
    },
    {
      scope: [
        "keyword",
        "keyword.function",
        "keyword.control",
        "keyword.other",
        "keyword.other.package",
        "keyword.other.import",
        "keyword.other.type",
      ],
      settings: { foreground: "#ff00ff" },
    },
    {
      scope: ["string", "string.quoted"],
      settings: { foreground: "#39ff14" },
    },
    {
      scope: ["constant.numeric", "constant.language"],
      settings: { foreground: "#39ff14" },
    },
    {
      scope: ["constant.other.placeholder"],
      settings: { foreground: "#00aaaa" },
    },
    {
      scope: ["constant.character.escape"],
      settings: { foreground: "#00aaaa" },
    },
    {
      scope: ["entity.name.tag"],
      settings: { foreground: "#00ffff" },
    },
    {
      scope: [
        "entity.name.function",
        "entity.name.function.component",
        "support.function.builtin",
      ],
      settings: { foreground: "#00ffff" },
    },
    {
      scope: ["entity.name.type"],
      settings: { foreground: "#00ffff" },
    },
    {
      scope: ["entity.other.attribute-name"],
      settings: { foreground: "#ff00ff" },
    },
    {
      scope: ["variable.parameter"],
      settings: { foreground: "#c0c8e0" },
    },
    {
      scope: ["punctuation.definition.component-call"],
      settings: { foreground: "#ff00ff" },
    },
    {
      scope: [
        "entity.name.function.component-call",
      ],
      settings: { foreground: "#00ffff" },
    },
    {
      scope: ["keyword.operator"],
      settings: { foreground: "#606888" },
    },
    {
      scope: ["variable.other"],
      settings: { foreground: "#c0c8e0" },
    },
    {
      scope: ["punctuation.definition.tag"],
      settings: { foreground: "#606888" },
    },
    {
      scope: ["punctuation.separator"],
      settings: { foreground: "#606888" },
    },
    {
      scope: ["punctuation.definition.block"],
      settings: { foreground: "#606888" },
    },
    {
      scope: ["entity.name.package"],
      settings: { foreground: "#c0c8e0" },
    },
    {
      scope: ["entity.name.import"],
      settings: { foreground: "#c0c8e0" },
    },
    {
      scope: ["variable.other.ref"],
      settings: { foreground: "#c0c8e0" },
    },
  ],
};

const lightTheme: ThemeRegistrationRaw = {
  name: "go-tui-light",
  type: "light",
  colors: {
    "editor.background": "#f8fafc",
    "editor.foreground": "#2a2a4e",
  },
  tokenColors: [
    {
      scope: ["comment", "comment.line", "comment.block"],
      settings: { foreground: "#8a8aaa" },
    },
    {
      scope: [
        "keyword",
        "keyword.function",
        "keyword.control",
        "keyword.other",
        "keyword.other.package",
        "keyword.other.import",
        "keyword.other.type",
      ],
      settings: { foreground: "#aa00aa" },
    },
    {
      scope: ["string", "string.quoted"],
      settings: { foreground: "#1a8a0a" },
    },
    {
      scope: ["constant.numeric", "constant.language"],
      settings: { foreground: "#1a8a0a" },
    },
    {
      scope: ["constant.other.placeholder"],
      settings: { foreground: "#006688" },
    },
    {
      scope: ["constant.character.escape"],
      settings: { foreground: "#006688" },
    },
    {
      scope: ["entity.name.tag"],
      settings: { foreground: "#0088aa" },
    },
    {
      scope: [
        "entity.name.function",
        "entity.name.function.component",
        "support.function.builtin",
      ],
      settings: { foreground: "#0088aa" },
    },
    {
      scope: ["entity.name.type"],
      settings: { foreground: "#0088aa" },
    },
    {
      scope: ["entity.other.attribute-name"],
      settings: { foreground: "#aa00aa" },
    },
    {
      scope: ["variable.parameter"],
      settings: { foreground: "#2a2a4e" },
    },
    {
      scope: ["punctuation.definition.component-call"],
      settings: { foreground: "#aa00aa" },
    },
    {
      scope: [
        "entity.name.function.component-call",
      ],
      settings: { foreground: "#0088aa" },
    },
    {
      scope: ["keyword.operator"],
      settings: { foreground: "#5a5a7e" },
    },
    {
      scope: ["variable.other"],
      settings: { foreground: "#2a2a4e" },
    },
    {
      scope: ["punctuation.definition.tag"],
      settings: { foreground: "#5a5a7e" },
    },
    {
      scope: ["punctuation.separator"],
      settings: { foreground: "#5a5a7e" },
    },
    {
      scope: ["punctuation.definition.block"],
      settings: { foreground: "#5a5a7e" },
    },
    {
      scope: ["entity.name.package"],
      settings: { foreground: "#2a2a4e" },
    },
    {
      scope: ["entity.name.import"],
      settings: { foreground: "#2a2a4e" },
    },
    {
      scope: ["variable.other.ref"],
      settings: { foreground: "#2a2a4e" },
    },
  ],
};

let highlighterInstance: HighlighterCore | null = null;
let highlighterPromise: Promise<HighlighterCore> | null = null;

export async function getHighlighter(): Promise<HighlighterCore> {
  if (highlighterInstance) return highlighterInstance;
  if (highlighterPromise) return highlighterPromise;

  highlighterPromise = createHighlighterCore({
    themes: [darkTheme, lightTheme],
    langs: [
      gsxGrammar as unknown as Parameters<HighlighterCore["loadLanguage"]>[0],
      goLang,
      shellLang,
    ],
    engine: createOnigurumaEngine(import("shiki/wasm")),
  }).then((h) => {
    highlighterInstance = h;
    return h;
  });

  return highlighterPromise;
}

export function highlight(
  code: string,
  lang: string,
  theme: "dark" | "light",
): string {
  if (!highlighterInstance) return "";
  const themeName = theme === "dark" ? "go-tui-dark" : "go-tui-light";
  const supportedLang =
    lang === "shell" || lang === "sh" || lang === "bash" ? "shellscript" : lang;

  try {
    return highlighterInstance.codeToHtml(code, {
      lang: supportedLang,
      theme: themeName,
    });
  } catch {
    return highlighterInstance.codeToHtml(code, {
      lang: "text",
      theme: themeName,
    });
  }
}
