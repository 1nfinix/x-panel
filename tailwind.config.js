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
        canvas: 'var(--xp-canvas)',
        'surface-soft': 'var(--xp-surface-soft)',
        'surface-card': 'var(--xp-surface-card)',
        'surface-strong': 'var(--xp-surface-strong)',
        hairline: 'var(--xp-hairline)',
        ink: 'var(--xp-ink)',
        body: 'var(--xp-body)',
        muted: 'var(--xp-muted)',
        'muted-soft': 'var(--xp-muted-soft)',
        primary: 'var(--xp-primary)',
        'primary-active': 'var(--xp-primary-active)',
      },
      borderRadius: {
        'xp-sm': 'var(--xp-radius-sm)',
        'xp-md': 'var(--xp-radius-md)',
        'xp-lg': 'var(--xp-radius-lg)',
        'xp-pill': 'var(--xp-radius-pill)',
      },
      fontFamily: {
        display: 'var(--xp-font-display)',
        body: 'var(--xp-font-body)',
        mono: 'var(--xp-font-mono)',
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
