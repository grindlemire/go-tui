export interface ContentPage {
  slug: string;
  title: string;
  order: number;
  body: string;
}

function parseFrontmatter(raw: string): { meta: Record<string, string>; body: string } {
  const match = raw.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n([\s\S]*)$/);
  if (!match) return { meta: {}, body: raw };

  const meta: Record<string, string> = {};
  for (const line of match[1].split("\n")) {
    const idx = line.indexOf(":");
    if (idx > 0) {
      const key = line.slice(0, idx).trim();
      let value = line.slice(idx + 1).trim();
      if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
        value = value.slice(1, -1);
      }
      meta[key] = value;
    }
  }
  return { meta, body: match[2] };
}

function loadPages(files: Record<string, string>): ContentPage[] {
  const pages: ContentPage[] = [];

  for (const [path, raw] of Object.entries(files)) {
    const filename = path.split("/").pop() ?? "";
    const { meta, body } = parseFrontmatter(raw);

    const orderMatch = filename.match(/^(\d+)-/);
    const order = meta.order
      ? parseInt(meta.order, 10)
      : orderMatch
        ? parseInt(orderMatch[1], 10)
        : 99;

    const slug = meta.slug ?? filename.replace(/^\d+-/, "").replace(/\.md$/, "");
    const title = meta.title ?? slug.replace(/-/g, " ");

    pages.push({ slug, title, order, body });
  }

  return pages.sort((a, b) => a.order - b.order);
}

const guideFiles = import.meta.glob("../content/guide/*.md", {
  query: "?raw",
  import: "default",
  eager: true,
}) as Record<string, string>;

const referenceFiles = import.meta.glob("../content/reference/*.md", {
  query: "?raw",
  import: "default",
  eager: true,
}) as Record<string, string>;

export function loadGuide(): ContentPage[] {
  return loadPages(guideFiles);
}

export function loadReference(): ContentPage[] {
  return loadPages(referenceFiles);
}
