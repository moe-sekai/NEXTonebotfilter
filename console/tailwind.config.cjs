/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          50: "#f7f6f2",
          100: "#ebe9e2",
          200: "#d6d2c4",
          300: "#aaa593",
          400: "#7e7a6b",
          500: "#5a5749",
          600: "#3d3b32",
          700: "#2a2922",
          800: "#1c1b16",
          900: "#0e0d0a",
        },
        clay: {
          400: "#d97757",
          500: "#cc785c",
          600: "#b45f44",
        },
      },
      fontFamily: {
        sans: ["ui-sans-serif", "system-ui", "Inter", "PingFang SC", "Microsoft YaHei", "sans-serif"],
        serif: ["Tiempos", "Source Serif", "Georgia", "serif"],
        mono: ["JetBrains Mono", "ui-monospace", "monospace"],
      },
      boxShadow: {
        soft: "0 1px 0 rgba(0,0,0,0.04), 0 1px 3px rgba(0,0,0,0.04)",
      },
    },
  },
  plugins: [],
};
