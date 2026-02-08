// Basic security utilities for markdown parsing

/**
 * Sanitize URLs to prevent XSS attacks
 * Removes dangerous schemes like javascript: and data:
 */
export function sanitizeURL(url) {
  if (!url) return "";
  const trimmed = url.trim();
  const lower = trimmed.toLowerCase();

  // Block javascript:, data:, and vbscript: schemes
  if (lower.startsWith("javascript:") || lower.startsWith("data:") || lower.startsWith("vbscript:")) {
    return "";
  }

  return trimmed;
}

/**
 * Checks if a URL is valid
 */
export function isValidURL(url) {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

/**
 * Escapes HTML special characters to prevent XSS
 */
export function escapeHTML(str) {
  if (!str) return "";
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#x27;")
    .replace(/\//g, "&#x2F;"); // optional, helps prevent closing tags
}
