---
title: Webauthn Browser APIs
---

This section is dedicated to the two most important JavaScript browser APIs in the Webauthn workflow, [`navigator.credentials.create`](https://developer.mozilla.org/en-US/docs/Web/API/CredentialsContainer/create) and [`navigator.credentials.get`](https://developer.mozilla.org/en-US/docs/Web/API/CredentialsContainer/get), their usage, configurations, and differences across browsers and platforms.
On top of that, you will learn about [`Uint8Array`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Uint8Array), [`atob`](https://developer.mozilla.org/en-US/docs/Web/API/atob), and URL-safe Base64 encoding.

All code snippets in this section will be using [TypeScript](https://www.typescriptlang.org/) syntax.
If you don't how to read and transform TypeScript code, you may wish to improve your Web development skills before implementing a Webauthn workflow on your website.

### Convert URL-safe Base64 to `Uint8Array`

```typescript
export const URLBase64ToUint8Array = (data: string): Uint8Array => {
  return Uint8Array.from(
    window.atob(data.replaceAll("-", "+").replaceAll("_", "/")),
    (c) => c.charCodeAt(0),
  );
};
```

### Convert an `ArrayBuffer` to URL-safe Base64

```typescript
export const bufferToURLBase64 = (data: ArrayBuffer): string => {
  const bytes = new Uint8Array(data);
  let str = "";
  for (const charCode of bytes) {
    str += String.fromCharCode(charCode);
  }
  return window.btoa(str)
    .replaceAll("+", "-")
    .replaceAll("/", "_")
    .replaceAll("=", "");
};
```
