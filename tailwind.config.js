/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        bg0: 'var(--x-bg0)',
        bg1: 'var(--x-bg1)',
        bg2: 'var(--x-bg2)',
        hover: 'var(--x-hover)',
        ph: 'var(--x-ph)',
        text1: 'var(--x-text1)',
        text2: 'var(--x-text2)',
        icon3: 'var(--x-icon3)',
        sep: 'var(--x-sep)',
        primary: 'var(--x-primary)',
        danger: 'var(--x-red-fg)',
      },
      borderRadius: {
        card: 'var(--x-r-card)',
        small: 'var(--x-r-small)',
        pill: 'var(--x-r-pill)',
        icon: 'var(--x-r-icon)',
        glyph: 'var(--x-r-glyph)',
      },
      fontFamily: {
        body: 'var(--x-font)',
      },
      animation: {
        blink: 'blink 1.2s infinite steps(1, start)',
      },
      keyframes: {
        blink: {
          '0%, 100%': { 'background-color': 'currentColor' },
          '50%': { 'background-color': 'transparent' },
        },
      },
    },
  },
  plugins: [],
}
