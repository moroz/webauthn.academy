document.querySelectorAll(".gist pre").forEach(o=>{const t=document.createElement("button");t.setAttribute("type","button"),t.className="button copy-to-clipboard",t.textContent="Copy to clipboard",t.addEventListener("click",()=>{let e=o.dataset.code?.split(`
`);if(!e)return;const l=o.querySelector("code.language-shell")!==null,a=e.some(n=>n?.startsWith("$"));l&&a&&(e=e.filter(n=>n?.startsWith("$")).map(n=>n.replace("$ ","")).filter(Boolean)),navigator.clipboard.writeText(e.join(`
`).trim())}),o.parentElement?.appendChild(t)});
