import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vitest/config"

// En desarrollo, /api se reenvía a la API local (puerto 5141), igual que el spa-server
// en Docker: el navegador ve un solo origen y la cookie de sesión funciona sin CORS.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5140,
    proxy: { "/api": "http://localhost:5141" },
  },
  test: {
    environment: "node",
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      reportsDirectory: "coverage",
    },
  },
})
