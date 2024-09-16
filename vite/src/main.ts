import "./main.scss";

document.querySelectorAll(".gist pre").forEach((pre) => {
  const button = document.createElement("button");
  button.setAttribute("type", "button");
  button.className = "button copy-to-clipboard";
  button.textContent = "Copy to clipboard";

  const lines = (pre as HTMLPreElement).dataset.code?.split("\n") as string[];
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
