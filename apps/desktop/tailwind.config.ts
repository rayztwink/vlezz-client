import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        graphite: {
          50: '#f8faf9',
          100: '#eef2f1',
          200: '#dbe3e1',
          500: '#64716f',
          700: '#33413f',
          900: '#111817'
        }
      },
      boxShadow: {
        panel: '0 1px 2px rgba(16, 24, 40, 0.05)'
      }
    }
  },
  plugins: []
} satisfies Config
