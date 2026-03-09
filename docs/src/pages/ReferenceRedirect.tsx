import { Navigate } from "react-router-dom";
import { loadReference } from "../lib/markdown.ts";

export default function ReferenceRedirect() {
  const pages = loadReference();
  if (pages.length === 0) return null;
  return <Navigate to={`/reference/${pages[0].slug}`} replace />;
}
