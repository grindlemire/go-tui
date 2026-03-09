import { useParams } from "react-router-dom";
import { loadGuide } from "../lib/markdown.ts";

export default function RawGuidePage() {
  const { slug } = useParams();
  const pages = loadGuide();
  const page = pages.find((p) => p.slug === slug);

  if (!page) return <pre>Guide not found.</pre>;

  return (
    <pre
      style={{
        margin: 0,
        padding: "1rem",
        whiteSpace: "pre-wrap",
        wordBreak: "break-word",
        fontFamily: "'Fira Code', monospace",
        fontSize: "13px",
        lineHeight: 1.6,
        background: "#1a1a2e",
        color: "#e0e0e0",
        minHeight: "100vh",
      }}
    >
      {page.body}
    </pre>
  );
}
