import { Marked } from 'marked';

const linkProtocols = new Set(['http:', 'https:', 'mailto:', 'tel:']);
const imageProtocols = new Set(['http:', 'https:']);

function escapeHtml(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

function escapeAttribute(value) {
  return escapeHtml(value).replaceAll('`', '&#96;');
}

function isSafeUrl(value, allowedProtocols) {
  const trimmed = [...value.trim()]
    .filter((char) => {
      const code = char.charCodeAt(0);
      return code > 0x20 && code !== 0x7f;
    })
    .join('');

  if (!trimmed) {
    return false;
  }

  if (trimmed.startsWith('#') || trimmed.startsWith('/') || trimmed.startsWith('./') || trimmed.startsWith('../')) {
    return true;
  }

  const scheme = /^[a-zA-Z][a-zA-Z0-9+.-]*:/.exec(trimmed);
  if (!scheme) {
    return true;
  }

  return allowedProtocols.has(scheme[0].toLowerCase());
}

const renderer = {
  html({ text }) {
    return escapeHtml(text);
  },
  link({ href, title, tokens }) {
    const label = this.parser.parseInline(tokens);

    if (!isSafeUrl(href, linkProtocols)) {
      return label;
    }

    const titleAttribute = title ? ` title="${escapeAttribute(title)}"` : '';
    return `<a href="${escapeAttribute(href)}"${titleAttribute}>${label}</a>`;
  },
  image({ href, title, text }) {
    if (!isSafeUrl(href, imageProtocols)) {
      return escapeHtml(text);
    }

    const titleAttribute = title ? ` title="${escapeAttribute(title)}"` : '';
    return `<img src="${escapeAttribute(href)}" alt="${escapeAttribute(text)}"${titleAttribute}>`;
  },
};

const markdown = new Marked({
  gfm: true,
  renderer,
});

export function renderHelpMarkdown(content) {
  return markdown.parse(content, { async: false });
}
