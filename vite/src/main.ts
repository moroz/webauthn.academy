import "./main.scss";

document.querySelectorAll("pre.chroma").forEach((block) => {
  const button = document.createElement("button");
  button.setAttribute("type", "button");
  button.className = "button copy-to-clipboard";
  button.textContent = "Copy to clipboard";

  button.addEventListener("click", () => {
    let lines = [...block.querySelectorAll(`span.cl`)]
      .map((line) => line.textContent)
      .join("")
      .trim();

    const isShell = block.querySelector("code.language-shell") !== null;

    if (isShell && lines.startsWith("$")) {
      lines = lines
        .split("\n")
        .filter((line) => line.startsWith("$"))
        .map((line) => line.replace("$ ", ""))
        .join("\n");
    }

    navigator.clipboard.writeText(lines);
  });

  const container = block.parentElement;
  container?.appendChild(button);
});
