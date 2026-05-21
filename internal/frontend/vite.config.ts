import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import basicSsl from "@vitejs/plugin-basic-ssl";

// https://vite.dev/config/
export default defineConfig(() => {
  return {
    plugins: [tanstackRouter({ target: "react" }), react(), basicSsl()],
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
        },
      },
    },
    base: "/ui/",
    build: {
      outDir: "build/ui",
    },
  };
});
