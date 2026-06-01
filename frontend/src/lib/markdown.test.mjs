import assert from 'node:assert/strict';
import test from 'node:test';

import { renderHelpMarkdown } from './markdown.js';

test('renderHelpMarkdown escapes raw script HTML', () => {
  const html = renderHelpMarkdown('# Title\n\n<script>alert("xss")</script>');

  assert.doesNotMatch(html, /<script/i);
  assert.match(html, /&lt;script&gt;alert\(&quot;xss&quot;\)&lt;\/script&gt;/);
});

test('renderHelpMarkdown escapes event handler attributes in raw HTML', () => {
  const html = renderHelpMarkdown('<img src=x onerror="alert(1)">');

  assert.doesNotMatch(html, /<img/i);
  assert.match(html, /&lt;img src=x onerror=&quot;alert\(1\)&quot;&gt;/);
});

test('renderHelpMarkdown removes javascript URLs from links and images', () => {
  const html = renderHelpMarkdown('[link](javascript:alert(1))\n\n![image](javascript:alert(1))');

  assert.doesNotMatch(html, /javascript:/i);
  assert.doesNotMatch(html, /href=/i);
  assert.doesNotMatch(html, /<img/i);
  assert.match(html, />link</);
  assert.match(html, />image</);
});

test('renderHelpMarkdown keeps normal markdown links and relative images', () => {
  const html = renderHelpMarkdown('[Porto](https://example.com "site")\n\n![logo](/images/logo.png)');

  assert.match(html, /<a href="https:\/\/example.com" title="site">Porto<\/a>/);
  assert.match(html, /<img src="\/images\/logo.png" alt="logo">/);
});
