import "./main.scss";
import Prism from "prismjs";
import "prismjs/plugins/line-numbers/prism-line-numbers.js";
import "prismjs/plugins/line-numbers/prism-line-numbers.css";

let code: any;

Prism.hooks.add("before-highlight", (env) => {
  code = env.element.innerHTML;
});

Prism.hooks.add("after-highlight", (env) => {
  env.element.innerHTML = code;
});

document.querySelectorAll(".gist pre").forEach((pre) => {
  const button = document.createElement("button");
  button.setAttribute("type", "button");
  button.className = "button copy-to-clipboard";
  button.textContent = "Copy to clipboard";

  let lines = (pre as HTMLPreElement).dataset.code?.split("\n") as string[];
  if (lines.length === 1) {
    button.classList.add("is-centered");
  }

  button.addEventListener("click", () => {
    let lines = (pre as HTMLPreElement).dataset.code?.split("\n") as string[];
    if (!lines) return;

    const isShell = pre.querySelector("code.language-shell") !== null;
    const hasDollar = lines.some((line) => line?.startsWith("$"));

    if (isShell && hasDollar) {
      lines = lines
        .filter((line) => line?.startsWith("$"))
        .map((line) => line!.replace("$ ", ""))
        .filter(Boolean);
    }

    navigator.clipboard.writeText(lines.join("\n").trim());
  });

  const container = pre.parentElement;
  container?.appendChild(button);
});
