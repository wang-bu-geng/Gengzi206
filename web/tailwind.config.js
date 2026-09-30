/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Primary brand colors - Apple Blue scale
        primary: {
          50: '#e8f2ff',
          100: '#cfe5ff',
          200: '#9fcbff',
          300: '#66abff',
          400: '#2e8dff',
          500: '#007aff',
          600: '#0064d6',
          700: '#004fad',
          800: '#00275a',
          900: '#001d47',
        },
        // Neutral colors for backgrounds
        surface: {
          light: '#ffffff',
          DEFAULT: '#f2f2f7',
          dark: '#000000',
        },
        // Card backgrounds
        card: {
          light: '#ffffff',
          dark: '#1c1c1e',
        },
        // Border colors
        border: {
          light: '#e5e5ea',
          dark: '#3a3a3c',
        },
        // Text colors
        content: {
          primary: '#1d1d1f',
          secondary: '#6e6e73',
          muted: '#aeaeb2',
          'primary-dark': '#f1f5f9',
          'secondary-dark': '#94a3b8',
          'muted-dark': '#64748b',
        },
      },
      boxShadow: {
        'card': '0 0.5px 1px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 2px 8px rgba(0, 0, 0, 0.08), 0 1px 3px rgba(0, 0, 0, 0.06)',
        'card-dark': '0 0.5px 1px rgba(0, 0, 0, 0.2), 0 1px 2px rgba(0, 0, 0, 0.3)',
        'card-hover-dark': '0 2px 8px rgba(0, 0, 0, 0.3), 0 1px 3px rgba(0, 0, 0, 0.2)',
      },
      borderRadius: {
        'xl': '0.875rem',
        '2xl': '1rem',
        '3xl': '1.5rem',
      },
      transitionTimingFunction: {
        'bounce-in': 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
      },
    },
  },
  plugins: [],
}
