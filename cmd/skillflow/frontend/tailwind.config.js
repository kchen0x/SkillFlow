/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        surface: 'var(--bg-surface)',
        elevated: 'var(--bg-elevated)',
        overlay: 'var(--bg-overlay)',
        accent: {
          DEFAULT: 'var(--accent-primary)',
          secondary: 'var(--accent-secondary)',
        },
        'tx-primary': 'var(--text-primary)',
        'tx-secondary': 'var(--text-secondary)',
        'tx-muted': 'var(--text-muted)',
        'btn-primary': 'var(--btn-primary-bg)',
      },
      boxShadow: {
        card: 'var(--shadow-card)',
        dialog: 'var(--shadow-dialog)',
        glow: 'var(--glow-accent)',
        'glow-sm': 'var(--glow-accent-sm)',
        'glow-btn': 'var(--glow-btn)',
      },
    },
  },
  plugins: [],
}
