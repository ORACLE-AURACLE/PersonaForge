import React from "react";
import { sanitizeURL, isValidURL, escapeHTML } from "./utils/security";

/**
 * Generates an anchor ID from heading text
 * Matches the format used in the privacy policy markdown (e.g., "1-privacy-policy", "11-information-we-collect")
 */
function generateAnchorId(text) {
  const cleanText = text
    .replace(/\*\*([^*]+)\*\*/g, "$1")
    .replace(/\*([^*]+)\*/g, "$1")
    .replace(/__([^_]+)__/g, "$1")
    .replace(/_([^_]+)_/g, "$1")
    .replace(/`([^`]+)`/g, "$1")
    .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1")
    .trim();

  const numberMatch = cleanText.match(/^(\d+(?:\.\d+)*)\.?\s*(.*)$/);
  let sectionNumber = "";
  let restOfText = cleanText;

  if (numberMatch) {
    sectionNumber = numberMatch[1].replace(/\./g, "");
    restOfText = numberMatch[2] || "";
  }

  const textPart = restOfText
    .toLowerCase()
    .replace(/[^\w\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");

  if (sectionNumber && textPart) return `${sectionNumber}-${textPart}`;
  if (sectionNumber) return sectionNumber;
  if (textPart) return textPart;

  return cleanText.toLowerCase().replace(/[^\w\s-]/g, "").replace(/\s+/g, "-").replace(/-+/g, "-").replace(/^-|-$/g, "");
}

export function parseMarkdown(content) {
  if (!content) return [];

  content = content.replace(/\r\n/g, "\n").replace(/\r/g, "\n");

  const result = [];
  const lines = content.split("\n");
  const context = {
    inCodeBlock: false,
    codeBlockLanguage: "",
    codeBlockContent: [],
  };

  let i = 0;
  while (i < lines.length) {
    const line = lines[i];

    // Code block handling removed for simplicity in chat/insights workflow

    // Horizontal rules
    if (/^[-*_]{3,}$/.test(line.trim())) {
      result.push(<hr key={`hr-${result.length}`} className="my-3 md:my-6 border-border" />);
      i++;
      continue;
    }

    // Headings
    const headingMatch = line.match(/^(#{1,6})\s+(.+)$/);
    if (headingMatch) {
      const level = headingMatch[1].length;
      const headingText = headingMatch[2];
      const text = parseInlineMarkdown(headingText, `heading-${result.length}`);
      const anchorId = generateAnchorId(headingText);

      const className =
        {
          1: "text-2xl md:text-4xl font-bold my-4 md:my-6 scroll-mt-20",
          2: "text-xl md:text-3xl font-bold my-3 md:my-5 scroll-mt-20",
          3: "text-lg md:text-2xl font-semibold my-2 md:my-4 scroll-mt-20",
          4: "text-base md:text-xl font-semibold my-2 md:my-3 scroll-mt-20",
          5: "text-sm md:text-lg font-semibold my-1 md:my-2 scroll-mt-20",
          6: "text-sm md:text-base font-semibold my-1 md:my-2 scroll-mt-20",
        }[level] || "text-sm md:text-base font-semibold my-1 md:my-2 scroll-mt-20";

      result.push(
        React.createElement(`h${level}`, { key: `heading-${result.length}`, id: anchorId, className }, text)
      );
      i++;
      continue;
    }

    // Blockquotes and tables removed for simplicity in chat/insights workflow

    // Lists
    if (/^(\s*)([-*+]|\d+\.)\s+/.test(line)) {
      const listResult = parseList(lines, i);
      if (listResult) {
        result.push(listResult.element);
        i = listResult.nextIndex;
        continue;
      }
    }

    // Empty lines
    if (line.trim() === "") {
      i++;
      continue;
    }

    // Paragraphs
    const paragraphLines = [];
    while (
      i < lines.length &&
      lines[i].trim() !== "" &&
      !lines[i].startsWith("#") &&
      !lines[i].startsWith("```") &&
      !lines[i].startsWith(">") &&
      !lines[i].includes("|") &&
      !/^(\s*)([-*+]|\d+\.)\s+/.test(lines[i]) &&
      !/^[-*_]{3,}$/.test(lines[i].trim())
    ) {
      paragraphLines.push(lines[i]);
      i++;
    }

    if (paragraphLines.length > 0) {
      const paragraphText = paragraphLines.join("\n");
      result.push(
        <p key={`p-${result.length}`} className="my-2 md:my-3 leading-relaxed">
          {parseInlineMarkdown(paragraphText, `p-${result.length}`)}
        </p>
      );
      continue;
    }

    i++;
  }

  return result;
}

// Inline Markdown parser
function parseInlineMarkdown(text, keyPrefix) {
  const nodes = [];
  let lastIndex = 0;
  let keyCounter = 0;

  const patterns = [
    // Code spans
    { regex: /`([^`]+)`/g, handler: (match) => <code key={`${keyPrefix}-code-${keyCounter++}`} className="bg-muted px-1.5 py-0.5 rounded text-sm font-mono">{match[1]}</code> },
    // Links
    {
      regex: /\[([^\]]+)\]\(([^)]+)\)/g,
      handler: (match) => {
        const rawUrl = match[2].trim();
        const isAnchorLink = rawUrl.startsWith("#");
        if (isAnchorLink) {
          const anchorId = rawUrl.slice(1);
          return <a key={`${keyPrefix}-link-${keyCounter++}`} href={`#${anchorId}`} onClick={(e) => {
            e.preventDefault();
            const element = document.getElementById(anchorId);
            if (element) element.scrollIntoView({ behavior: "smooth", block: "start" });
            window.history.pushState(null, "", `#${anchorId}`);
          }} className="text-primary underline hover:text-primary/80">{parseInlineMarkdown(match[1], `${keyPrefix}-link-${keyCounter}`)}</a>;
        }
        const url = sanitizeURL(rawUrl);
        if (!url || !isValidURL(url)) return <span key={`${keyPrefix}-link-${keyCounter++}`}>{parseInlineMarkdown(match[1], `${keyPrefix}-link-${keyCounter}`)}</span>;
        return <a key={`${keyPrefix}-link-${keyCounter++}`} href={url} target="_blank" rel="noopener noreferrer" className="text-primary underline hover:text-primary/80">{parseInlineMarkdown(match[1], `${keyPrefix}-link-${keyCounter}`)}</a>;
      }
    },
    // Images
    {
      regex: /!\[([^\]]*)\]\(([^)]+)\)/g,
      handler: (match) => {
        const url = sanitizeURL(match[2]);
        if (!url || !isValidURL(url)) return <React.Fragment key={`${keyPrefix}-img-${keyCounter++}`} />;
        return <img key={`${keyPrefix}-img-${keyCounter++}`} src={url} alt={escapeHTML(match[1] || "")} className="max-w-full h-auto my-4 rounded-lg" loading="lazy" />;
      }
    },
    // Bold
    { regex: /(\*\*|__)([^*_\n]+?)\1/g, handler: (match) => <strong key={`${keyPrefix}-bold-${keyCounter++}`} className="font-bold">{parseInlineMarkdown(match[2], `${keyPrefix}-bold-${keyCounter}`)}</strong> },
    // Italic
    { regex: /(?<!\*)\*(?!\*)([^*\n]+?)\*(?!\*)|(?<!_)_(?!_)([^_\n]+?)_(?!_)/g, handler: (match) => <em key={`${keyPrefix}-italic-${keyCounter++}`} className="italic">{parseInlineMarkdown(match[1] || match[2], `${keyPrefix}-italic-${keyCounter}`)}</em> },
    // Strikethrough
    { regex: /~~([^~]+?)~~/g, handler: (match) => <del key={`${keyPrefix}-del-${keyCounter++}`} className="line-through">{match[1]}</del> },
  ];

  const allMatches = [];
  patterns.forEach(({ regex, handler }) => {
    let match;
    regex.lastIndex = 0;
    while ((match = regex.exec(text)) !== null) allMatches.push({ index: match.index, length: match[0].length, handler, match });
  });

  allMatches.sort((a, b) => a.index - b.index);

  const nonOverlapping = [];
  for (const match of allMatches) {
    const overlaps = nonOverlapping.some(existing => match.index < existing.index + existing.length && match.index + match.length > existing.index);
    if (!overlaps) nonOverlapping.push(match);
  }

  for (const { index, length, handler, match } of nonOverlapping) {
    if (index > lastIndex) nodes.push(text.slice(lastIndex, index));
    nodes.push(handler(match));
    lastIndex = index + length;
  }

  if (lastIndex < text.length) nodes.push(text.slice(lastIndex));
  return nodes.length > 0 ? nodes : [text];
}

// Table parser removed for simplicity in chat/insights workflow

// List parser
function parseList(lines, startIndex) {
  const items = [];
  let i = startIndex;
  const stack = [{ level: -1, items }];

  while (i < lines.length) {
    const line = lines[i];
    const listMatch = line.match(/^(\s*)([-*+]|\d+\.)\s+(.+)$/);
    if (!listMatch) {
      if (line.trim() && /^\s+/.test(line) && stack.length > 1) {
        const lastItem = stack[stack.length - 1].items[stack[stack.length - 1].items.length - 1];
        if (lastItem) lastItem.content += " " + line.trim();
        i++;
        continue;
      }
      break;
    }

    const indent = listMatch[1].length;
    const content = listMatch[3];
    const level = Math.floor(indent / 2);

    while (stack.length > 1 && stack[stack.length - 1].level >= level) stack.pop();

    const currentList = stack[stack.length - 1].items;
    const item = { content, children: [] };
    currentList.push(item);
    stack.push({ level, items: item.children });
    i++;
  }

  if (!items.length) return null;

  const firstLine = lines[startIndex];
  const firstMatch = firstLine.match(/^(\s*)([-*+]|\d+\.)\s+/);
  const isOrdered = firstMatch ? /^\d+\.$/.test(firstMatch[2]) : false;

  const renderList = (listItems, ordered, depth = 0) => {
    const ListTag = ordered ? "ol" : "ul";
    const className = ordered ? "list-decimal list-inside my-2 space-y-1 ml-4" : "list-disc list-inside my-2 space-y-1 ml-4";
    return (
      <ListTag key={`list-${depth}-${startIndex}`} className={className}>
        {listItems.map((item, idx) => (
          <li key={`li-${depth}-${idx}`} className="my-1">
            {parseInlineMarkdown(item.content, `list-${depth}-${idx}`)}
            {item.children.length > 0 && <div className="ml-4 mt-1">{renderList(item.children, ordered, depth + 1)}</div>}
          </li>
        ))}
      </ListTag>
    );
  };

  return { element: renderList(items, isOrdered), nextIndex: i };
}
