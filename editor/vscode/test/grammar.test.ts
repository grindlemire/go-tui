import { describe, expect, test } from "bun:test";
import fs from "node:fs";
import path from "node:path";
import * as oniguruma from "vscode-oniguruma";
import * as vsctm from "vscode-textmate";

// GRAMMAR_PATH lets the same test run against another grammar file, which is
// how a change is checked to fail against the previous version.
const grammarPath = process.env.GRAMMAR_PATH ?? path.join(import.meta.dir, "..", "syntaxes", "gsx.tmLanguage.json");

const wasm = fs.readFileSync(path.join(path.dirname(require.resolve("vscode-oniguruma")), "onig.wasm")).buffer;
const onigLib = oniguruma.loadWASM(wasm).then(() => ({
  createOnigScanner: (patterns: string[]) => new oniguruma.OnigScanner(patterns),
  createOnigString: (s: string) => new oniguruma.OnigString(s),
}));
const registry = new vsctm.Registry({
  onigLib,
  loadGrammar: async () => vsctm.parseRawGrammar(fs.readFileSync(grammarPath, "utf8"), grammarPath),
});

type Token = { text: string; scopes: string[] };

async function tokenize(source: string): Promise<Token[][]> {
  const grammar = await registry.loadGrammar("source.gsx");
  if (!grammar) throw new Error("grammar failed to load");
  let state = vsctm.INITIAL;
  return source.split("\n").map((line) => {
    const r = grammar.tokenizeLine(line, state);
    state = r.ruleStack;
    return r.tokens.map((t) => ({ text: line.slice(t.startIndex, t.endIndex), scopes: t.scopes }));
  });
}

// innermost returns the last scope of the token whose text equals `text` on the given line.
function innermost(lines: Token[][], line: number, text: string): string {
  const tok = lines[line].find((t) => t.text === text);
  if (!tok) throw new Error(`no token ${JSON.stringify(text)} on line ${line}: ${lines[line].map((t) => t.text).join("|")}`);
  return tok.scopes[tok.scopes.length - 1];
}

const source = `package main

templ App(items []int) {
	<div class="flex-col">
		@widgets.Header("hover or F12")
		@Card(42) {
			<span>@c.icon (beta)</span>
		}
		@widgets.Row(fmt.Sprintf("%d items", len(items)), true)
		h := @widgets.Header("bound")
		@widgets.Multi(
			"first",
			2,
		)
	</div>
}`;

const CALL = "entity.name.function.component-call.gsx";

describe("component calls", () => {
  test("qualified and bare names are function-scoped as a whole", async () => {
    const lines = await tokenize(source);
    expect(innermost(lines, 4, "widgets.Header")).toBe(CALL);
    expect(innermost(lines, 5, "Card")).toBe(CALL);
    expect(innermost(lines, 8, "widgets.Row")).toBe(CALL);
  });

  test("arguments are highlighted as Go", async () => {
    const lines = await tokenize(source);
    expect(innermost(lines, 4, "hover or F12")).toBe("string.quoted.double.gsx");
    expect(innermost(lines, 5, "42")).toBe("constant.numeric.integer.gsx");
    expect(innermost(lines, 8, "Sprintf")).toBe("entity.name.function.go.gsx");
    expect(innermost(lines, 8, "%d")).toBe("constant.other.placeholder.gsx");
    expect(innermost(lines, 8, "len")).toBe("support.function.builtin.go.gsx");
    expect(innermost(lines, 8, "true")).toBe("constant.language.go.gsx");
  });

  test("multi-line arguments stay inside the call", async () => {
    const lines = await tokenize(source);
    expect(innermost(lines, 11, "first")).toBe("string.quoted.double.gsx");
    expect(innermost(lines, 12, "2")).toBe("constant.numeric.integer.gsx");
    expect(innermost(lines, 13, ")")).toBe("punctuation.definition.arguments.end.gsx");
  });

  test("a qualified call in a short binding keeps the binding scope", async () => {
    const lines = await tokenize(source);
    expect(innermost(lines, 9, "h")).toBe("variable.other.gsx");
    expect(innermost(lines, 9, "widgets.Header")).toBe(CALL);
  });

  test("an expression followed by spaced paren text is not a call", async () => {
    const lines = await tokenize(source);
    // The lexer needs the paren adjacent; "@c.icon (beta)" is an expression plus text.
    expect(innermost(lines, 6, "@c.icon (beta)")).toBe("meta.component.gsx");
  });
});
