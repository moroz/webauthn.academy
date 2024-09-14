document.querySelectorAll(".gist pre").forEach(l=>{const t=document.createElement("button");t.setAttribute("type","button"),t.className="button copy-to-clipboard",t.textContent="Copy to clipboard",(l.dataset.code?.split(`
`)).length===1&&t.classList.add("is-centered"),t.addEventListener("click",()=>{let e=l.dataset.code?.split(`
`);if(!e)return;const o=l.querySelector("code.language-shell")!==null,a=e.some(n=>n?.startsWith("$"));o&&a&&(e=e.filter(n=>n?.startsWith("$")).map(n=>n.replace("$ ","")).filter(Boolean)),navigator.clipboard.writeText(e.join(`
`).trim())}),l.parentElement?.appendChild(t)});
