import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { scopoIT } from "../content/scopo.it";
import { scopoEN } from "../content/scopo.en";

/**
 * /scopo — page describing what Pesa is for, what it is NOT for, and sources.
 * Rendered from the checked-in canonical markdown (see `src/content/scopo.*.ts`).
 */
export default function ScopePage() {
  const { i18n } = useTranslation();
  const markdown = i18n.language.startsWith("en") ? scopoEN : scopoIT;
  const html = useMemo(() => mdToHtml(markdown), [markdown]);

  return (
    <article
      className="prose surface mx-auto max-w-3xl p-6 sm:p-10 rise"
      // Content is trusted: checked-in source, not user input.
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

/**
 * Tiny markdown-to-HTML converter for our controlled content set.
 * Handles: H1-H3, bold/italic, links, unordered lists, paragraphs, inline code.
 * Deliberately minimal — no external dependency. A vite-plugin-md swap is planned.
 */
function mdToHtml(md: string): string {
  const esc = (s: string) =>
    s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

  const lines = md.split(/\r?\n/);
  const out: string[] = [];
  let inList = false;
  let paragraph: string[] = [];

  const flushPara = () => {
    if (paragraph.length) {
      out.push(`<p>${inline(paragraph.join(" "))}</p>`);
      paragraph = [];
    }
  };
  const closeList = () => {
    if (inList) {
      out.push("</ul>");
      inList = false;
    }
  };

  function inline(s: string): string {
    let t = esc(s);
    t = t.replace(/`([^`]+)`/g, "<code>$1</code>");
    t = t.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
    t = t.replace(/(?<!\*)\*([^*]+)\*(?!\*)/g, "<em>$1</em>");
    t = t.replace(
      /\[([^\]]+)\]\(([^)]+)\)/g,
      (_m, text, href) =>
        `<a href="${String(href).replace(/"/g, "&quot;")}" target="_blank" rel="noopener noreferrer">${text}</a>`,
    );
    return t;
  }

  for (const raw of lines) {
    const line = raw.trimEnd();
    if (/^\s*$/.test(line)) {
      flushPara();
      closeList();
      continue;
    }
    if (line.startsWith("# ")) {
      flushPara();
      closeList();
      out.push(`<h1>${inline(line.slice(2))}</h1>`);
      continue;
    }
    if (line.startsWith("## ")) {
      flushPara();
      closeList();
      out.push(`<h2>${inline(line.slice(3))}</h2>`);
      continue;
    }
    if (line.startsWith("### ")) {
      flushPara();
      closeList();
      out.push(`<h3>${inline(line.slice(4))}</h3>`);
      continue;
    }
    if (line.startsWith("---")) {
      flushPara();
      closeList();
      out.push("<hr />");
      continue;
    }
    if (line.startsWith("> ")) {
      flushPara();
      closeList();
      out.push(`<blockquote>${inline(line.slice(2))}</blockquote>`);
      continue;
    }
    if (/^[-*] /.test(line)) {
      flushPara();
      if (!inList) {
        out.push("<ul>");
        inList = true;
      }
      out.push(`<li>${inline(line.replace(/^[-*] /, ""))}</li>`);
      continue;
    }
    paragraph.push(line);
  }
  flushPara();
  closeList();
  return out.join("\n");
}
