import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import type { ThemeRegistrationRaw, LanguageInput } from "shiki/core";
import { createOnigurumaEngine } from "shiki/engine/oniguruma";
import goLang from "@shikijs/langs/go";
import shellLang from "@shikijs/langs/shellscript";
import gsxGrammar from "../../../editor/vscode/syntaxes/gsx.tmLanguage.json";

const darkTokenColors = [
    {
      scope: ["comment", "comment.line", "comment.block"],
      settings: { foreground: "#75715e" },
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
      settings: { foreground: "#f92672" },
    },
    {
      scope: ["string", "string.quoted"],
      settings: { foreground: "#e6db74" },
    },
    {
      scope: ["constant.numeric", "constant.language"],
      settings: { foreground: "#ae81ff" },
    },
    {
      scope: ["constant.other.placeholder"],
      settings: { foreground: "#ae81ff" },
    },
    {
      scope: ["constant.character.escape"],
      settings: { foreground: "#ae81ff" },
    },
    {
      scope: ["entity.name.tag"],
      settings: { foreground: "#f92672" },
    },
    {
      scope: [
        "entity.name.function",
        "entity.name.function.component",
        "support.function.builtin",
      ],
      settings: { foreground: "#a6e22e" },
    },
    {
      scope: ["entity.name.type"],
      settings: { foreground: "#66d9ef" },
    },
    {
      scope: ["entity.other.attribute-name"],
      settings: { foreground: "#a6e22e" },
    },
    {
      scope: ["variable.parameter"],
      settings: { foreground: "#fd971f" },
    },
    {
      scope: ["punctuation.definition.component-call"],
      settings: { foreground: "#f92672" },
    },
    {
      scope: [
        "entity.name.function.component-call",
      ],
      settings: { foreground: "#a6e22e" },
    },
    {
      scope: ["keyword.operator"],
      settings: { foreground: "#f92672" },
    },
    {
      scope: ["variable.other"],
      settings: { foreground: "#f8f8f2" },
    },
    {
      scope: ["punctuation.definition.tag"],
      settings: { foreground: "#f8f8f2" },
    },
    {
      scope: ["punctuation.separator"],
      settings: { foreground: "#f8f8f2" },
    },
    {
      scope: ["punctuation.definition.block"],
      settings: { foreground: "#f8f8f2" },
    },
    {
      scope: ["entity.name.package"],
      settings: { foreground: "#f8f8f2" },
    },
    {
      scope: ["variable.other.ref"],
      settings: { foreground: "#f8f8f2" },
    },
];

const darkTheme: ThemeRegistrationRaw = {
  name: "go-tui-dark",
  type: "dark",
  settings: darkTokenColors,
  colors: {
    "editor.background": "#23241e",
    "editor.foreground": "#f8f8f2",
  },
  tokenColors: darkTokenColors,
};

const lightTokenColors = [
    {
      scope: ["comment", "comment.line", "comment.block"],
      settings: { foreground: "#767260" },
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
      settings: { foreground: "#c01f5c" },
    },
    {
      scope: ["string", "string.quoted"],
      settings: { foreground: "#7a6e00" },
    },
    {
      scope: ["constant.numeric", "constant.language"],
      settings: { foreground: "#6e5dc6" },
    },
    {
      scope: ["constant.other.placeholder"],
      settings: { foreground: "#6e5dc6" },
    },
    {
      scope: ["constant.character.escape"],
      settings: { foreground: "#6e5dc6" },
    },
    {
      scope: ["entity.name.tag"],
      settings: { foreground: "#c01f5c" },
    },
    {
      scope: [
        "entity.name.function",
        "entity.name.function.component",
        "support.function.builtin",
      ],
      settings: { foreground: "#507009" },
    },
    {
      scope: ["entity.name.type"],
      settings: { foreground: "#217f96" },
    },
    {
      scope: ["entity.other.attribute-name"],
      settings: { foreground: "#507009" },
    },
    {
      scope: ["variable.parameter"],
      settings: { foreground: "#9a4c0b" },
    },
    {
      scope: ["punctuation.definition.component-call"],
      settings: { foreground: "#c01f5c" },
    },
    {
      scope: [
        "entity.name.function.component-call",
      ],
      settings: { foreground: "#507009" },
    },
    {
      scope: ["keyword.operator"],
      settings: { foreground: "#c01f5c" },
    },
    {
      scope: ["variable.other"],
      settings: { foreground: "#3d3c34" },
    },
    {
      scope: ["punctuation.definition.tag"],
      settings: { foreground: "#3d3c34" },
    },
    {
      scope: ["punctuation.separator"],
      settings: { foreground: "#3d3c34" },
    },
    {
      scope: ["punctuation.definition.block"],
      settings: { foreground: "#3d3c34" },
    },
    {
      scope: ["entity.name.package"],
      settings: { foreground: "#3d3c34" },
    },
    {
      scope: ["variable.other.ref"],
      settings: { foreground: "#3d3c34" },
    },
];

const lightTheme: ThemeRegistrationRaw = {
  name: "go-tui-light",
  type: "light",
  settings: lightTokenColors,
  colors: {
    "editor.background": "#f5f5f1",
    "editor.foreground": "#3d3c34",
  },
  tokenColors: lightTokenColors,
};

let highlighterInstance: HighlighterCore | null = null;
let highlighterPromise: Promise<HighlighterCore> | null = null;

export async function getHighlighter(): Promise<HighlighterCore> {
  if (highlighterInstance) return highlighterInstance;
  if (highlighterPromise) return highlighterPromise;

  highlighterPromise = createHighlighterCore({
    themes: [darkTheme, lightTheme],
    langs: [
      {
        ...gsxGrammar,
        name: "gsx",
      } as unknown as LanguageInput,
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
