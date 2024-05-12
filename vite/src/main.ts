import "./main.scss";

document.querySelectorAll("pre.chroma").forEach((block) => {
  const button = document.createElement("button");
  button.setAttribute("type", "button");
  button.className = "button copy-to-clipboard";
  button.textContent = "Copy to clipboard";

  button.addEventListener("click", () => {
    const lines = [...block.querySelectorAll(`span.cl`)]
      .map((line) => line.textContent)
      .join("")
      .trim();

    navigator.clipboard.writeText(lines);
  });

  const container = block.parentElement;
  container?.appendChild(button);
});
