import MiniSearch, { type SearchResult } from "minisearch";
import { loadGuide, loadReference, type ContentPage } from "./markdown";
import { extractHeadingIds } from "../components/Markdown.tsx";

export interface SearchHit {
  /** Navigation path, e.g. "guide/getting-started" or "reference/element" */
  path: string;
  /** Hash fragment for deep linking, e.g. "flexbox-layout" (empty for page-level matches) */
  hash: string;
  /** Page title */
  pageTitle: string;
  /** Section heading text (empty for page-level matches) */
  sectionTitle: string;
  type: "guide" | "reference";
  snippet: string;
}

function stripMarkdown(md: string): string {
  return md
    .replace(/```[\s\S]*?```/g, " ")       // fenced code blocks
    .replace(/`[^`]+`/g, " ")              // inline code
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, "$1") // links/images
    .replace(/#{1,6}\s+/g, "")             // headings
    .replace(/[*_~]{1,3}/g, "")            // bold/italic/strikethrough
    .replace(/^\s*[-*+]\s+/gm, "")         // unordered list markers
    .replace(/^\s*\d+\.\s+/gm, "")         // ordered list markers
    .replace(/^\s*>\s+/gm, "")             // blockquotes
    .replace(/\|/g, " ")                   // table pipes
    .replace(/---+/g, "")                  // horizontal rules
    .replace(/\n{2,}/g, "\n")
    .trim();
}

function extractSnippet(body: string, terms: string[]): string {
  const lower = body.toLowerCase();

  let best = -1;
  for (const term of terms) {
    const idx = lower.indexOf(term.toLowerCase());
    if (idx !== -1 && (best === -1 || idx < best)) {
      best = idx;
    }
  }

  if (best === -1) return body.slice(0, 120) + (body.length > 120 ? "..." : "");

  const start = Math.max(0, best - 40);
  const end = Math.min(body.length, best + 80);
  let snippet = body.slice(start, end).replace(/\s+/g, " ").trim();
  if (start > 0) snippet = "..." + snippet;
  if (end < body.length) snippet = snippet + "...";
  return snippet;
}

/** Split a page's markdown into sections keyed by heading ID. */
function splitSections(page: ContentPage, type: "guide" | "reference"): SearchDoc[] {
  const headings = extractHeadingIds(page.body);
  const lines = page.body.split("\n");
  const basePath = `${type}/${page.slug}`;
  const results: SearchDoc[] = [];

  // Content before the first heading → page-level doc
  const firstHeadingLine = headings.length > 0 ? headings[0].line - 1 : lines.length;
  const introText = stripMarkdown(lines.slice(0, firstHeadingLine).join("\n"));
  if (introText.length > 20) {
    results.push({
      id: `${basePath}::intro`,
      path: basePath,
      hash: "",
      pageTitle: page.title,
      sectionTitle: "",
      body: introText,
      type,
    });
  }

  // Each heading section
  for (let i = 0; i < headings.length; i++) {
    const startLine = headings[i].line - 1; // 1-indexed → 0-indexed
    const endLine = i + 1 < headings.length ? headings[i + 1].line - 1 : lines.length;
    const sectionText = stripMarkdown(lines.slice(startLine, endLine).join("\n"));

    results.push({
      id: `${basePath}::${headings[i].id}`,
      path: basePath,
      hash: headings[i].id,
      pageTitle: page.title,
      sectionTitle: headings[i].text,
      body: sectionText,
      type,
    });
  }

  // If no headings at all, index the whole page
  if (results.length === 0) {
    results.push({
      id: `${basePath}::full`,
      path: basePath,
      hash: "",
      pageTitle: page.title,
      sectionTitle: "",
      body: stripMarkdown(page.body),
      type,
    });
  }

  return results;
}

interface SearchDoc {
  id: string;
  path: string;
  hash: string;
  pageTitle: string;
  sectionTitle: string;
  body: string;
  type: "guide" | "reference";
}

let index: MiniSearch<SearchDoc> | null = null;
let docs: SearchDoc[] = [];

function ensureIndex(): MiniSearch<SearchDoc> {
  if (index) return index;

  const guides = loadGuide();
  const refs = loadReference();

  docs = [
    ...guides.flatMap((p) => splitSections(p, "guide")),
    ...refs.flatMap((p) => splitSections(p, "reference")),
  ];

  index = new MiniSearch<SearchDoc>({
    fields: ["pageTitle", "sectionTitle", "body"],
    storeFields: ["path", "hash", "pageTitle", "sectionTitle", "type"],
    searchOptions: {
      boost: { sectionTitle: 4, pageTitle: 3 },
      prefix: true,
      fuzzy: 0.2,
    },
  });

  index.addAll(docs);
  return index;
}

export function search(query: string): SearchHit[] {
  if (!query.trim()) return [];

  const idx = ensureIndex();
  const results: SearchResult[] = idx.search(query);

  const terms = query.trim().split(/\s+/);

  // Deduplicate: keep only the best-scoring result per path+hash
  const seen = new Set<string>();

  return results
    .filter((r) => {
      const key = `${r.path}#${r.hash}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    })
    .slice(0, 12)
    .map((r) => {
      const doc = docs.find((d) => d.id === r.id);
      return {
        path: (r.path ?? doc?.path ?? "") as string,
        hash: (r.hash ?? doc?.hash ?? "") as string,
        pageTitle: (r.pageTitle ?? doc?.pageTitle ?? "") as string,
        sectionTitle: (r.sectionTitle ?? doc?.sectionTitle ?? "") as string,
        type: (r.type ?? doc?.type ?? "guide") as "guide" | "reference",
        snippet: doc ? extractSnippet(doc.body, terms) : "",
      };
    });
}
