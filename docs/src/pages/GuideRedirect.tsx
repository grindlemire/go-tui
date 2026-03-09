import { Navigate } from "react-router-dom";
import { loadGuide } from "../lib/markdown.ts";

export default function GuideRedirect() {
  const pages = loadGuide();
  if (pages.length === 0) return null;
  return <Navigate to={`/guide/${pages[0].slug}`} replace />;
}
