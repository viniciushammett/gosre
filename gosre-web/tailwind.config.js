/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        surface: {
          DEFAULT: "#0f1117",
          raised: "#171b26",
          border: "#1e2233",
        },
        brand: {
          DEFAULT: "#4f8ef7",
          dim: "#2d5bbf",
        },
        status: {
          ok: "#22c55e",
          fail: "#ef4444",
          timeout: "#f59e0b",
          unknown: "#6b7280",
        },
      },
      fontFamily: {
        sans: ["IBM Plex Sans", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["IBM Plex Mono", "ui-monospace", "monospace"],
      },
    },
  },
  plugins: [],
};
