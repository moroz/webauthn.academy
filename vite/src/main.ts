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

    const isShell =
      pre.querySelector("code.language-shell, code.language-plain") !== null;
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

let contentBottom: number;

function setUpScrollEvent() {
  const toc = document.querySelector("aside.toc");
  const container = toc?.parentNode as HTMLDivElement;
  const footer = document.querySelector("footer");
  if (!container || !footer) return;

  const tocContent = toc!.querySelector(".wrapper");
  if (!tocContent) return;

  const rect = tocContent.getBoundingClientRect();
  contentBottom = rect.bottom;

  if (rect.height < 300) return;

  function handleScroll() {
    toc?.classList.toggle(
      "stick-to-bottom",
      contentBottom > container.getBoundingClientRect().bottom,
    );
  }

  const observer = new IntersectionObserver(console.log, {});
  observer.observe(footer);

  window.addEventListener("scroll", handleScroll);
}

window.addEventListener("resize", () => {
  const toc = document.querySelector("aside.toc");
  const tocContent = toc!.querySelector(".wrapper");
  if (!tocContent) return;
  contentBottom = tocContent?.getBoundingClientRect().bottom;
});

setUpScrollEvent();
