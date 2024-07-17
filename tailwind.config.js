/** @type {import('tailwindcss').Config} */
export default {
    content: [
      './static/**/*.{html,js}',
      './internal/**/*.{html,js}',
      './*.go',
    ],
    theme: {
      extend: {
        colors: {
          'yellowBackGround': '#FEF72D',
        },
        wave: {
            "0%": { transform: "rotate(0.0deg)" },
            "10%": { transform: "rotate(14deg)" },
            "20%": { transform: "rotate(-8deg)" },
            "30%": { transform: "rotate(14deg)" },
            "40%": { transform: "rotate(-4deg)" },
            "50%": { transform: "rotate(10.0deg)" },
            "60%": { transform: "rotate(0.0deg)" },
            "100%": { transform: "rotate(0.0deg)" },
        },
      },
    },
    plugins: [],
  }
  