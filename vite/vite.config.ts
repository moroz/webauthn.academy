import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// https://vitejs.dev/config/
export default defineConfig(({ command }) => {
  const isDev = command !== "build";

  return {
    plugins: [svelte()],
    build: {
      target: "esnext", // build for recent browsers
      outDir: "../assets", // emit assets to priv/static
      emptyOutDir: true,
      sourcemap: isDev, // enable source map in dev build
      manifest: false, // do not generate manifest.json
      rollupOptions: {
        input: {
          main: "./src/main.ts",
        },
        output: {
          entryFileNames: "[name].js", // remove hash
          chunkFileNames: "[name].js",
          assetFileNames: "[name][extname]",
        },
      },
    },
  };
});
