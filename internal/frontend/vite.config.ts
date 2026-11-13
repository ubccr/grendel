import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import basicSsl from "@vitejs/plugin-basic-ssl";

// https://vite.dev/config/
export default defineConfig(() => {
  return {
    plugins: [tanstackRouter({ target: "react", autoCodeSplitting: true }), react(), basicSsl()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      proxy: {
        "/v1": {
          target: "https://127.0.0.1:8080",
          secure: false,
          ws: true,
        },
      },
    },
    base: "/ui/",
    build: {
      // built straight into the directory grendeld embeds; it sits outside the Vite root, so emptying it has to be asked for, and that only clears ui/, leaving the committed .gitkeep beside it
      outDir: "../api/build/ui",
      emptyOutDir: true,
      // routes are split, so this only guards the entry chunk, about 630 kB with React, the router, Base UI and the shell; a vendor chunk was measured and made the first load larger
      chunkSizeWarningLimit: 700,
    },
  };
});
