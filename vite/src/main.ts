import "./main.scss";

document.querySelectorAll("pre.chroma").forEach((block) => {
  const button = document.createElement("button");
  button.setAttribute("type", "button");
  button.className = "button copy-to-clipboard";
  button.textContent = "Copy to clipboard";

  button.addEventListener("click", () => {
    let lines = [...block.querySelectorAll(`span.cl`)].map(
      (line) => line.textContent,
    );

    const isShell = block.querySelector("code.language-shell") !== null;
    const hasDollar = lines.some((line) => line?.startsWith("$"));

    if (isShell && hasDollar) {
      lines = lines
        .filter((line) => line?.startsWith("$"))
        .map((line) => line!.replace("$ ", "").trim())
        .filter(Boolean);
    }

    navigator.clipboard.writeText(lines.join("\n").trim());
  });

  const container = block.parentElement;
  container?.appendChild(button);
});
