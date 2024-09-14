import { DOMParser } from "https://deno.land/x/deno_dom/deno-dom-wasm.ts";
import { minify } from "https://esm.sh/html-minifier-terser@7.2.0";
import "./src/prism.js";
import fs from "node:fs/promises";
import { glob } from "https://esm.sh/glob@11.0.0";

import Prism from "prismjs";

const files = await glob("public/**/*.html");

for (const file of files) {
  const markup = await fs.readFile(file);
  const doc = new DOMParser().parseFromString(markup.toString(), "text/html");
  globalThis.document = doc;
  Prism.highlightAllUnder(doc);
  const html = doc.documentElement.outerHTML;
  const minified = await minify(html, {
    collapseWhitespace: true,
    collapseBooleanAttributes: true,
    decodeEntities: true,
    removeAttributeQuotes: true,
  });
  const combined = "<!DOCTYPE html>" + minified;
  await fs.writeFile(file, combined);
}
