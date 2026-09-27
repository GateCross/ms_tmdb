/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{vue,ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#171717",
        haze: "#f8f6f2",
        sand: "#e7dcc8",
        coral: "#f47c64",
        pine: "#275a53",
      },
      boxShadow: {
        soft: "0 20px 45px rgba(23, 23, 23, 0.12)",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};
