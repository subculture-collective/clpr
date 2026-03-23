import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    // Mobile-first responsive breakpoints
    screens: {
      xs: '375px',
      sm: '640px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
      '2xl': '1536px',
    },
    extend: {
      colors: {
        // === New design system tokens ===

        // Surfaces
        surface: {
          DEFAULT: 'rgb(var(--color-surface) / <alpha-value>)',
          raised: 'rgb(var(--color-surface-raised) / <alpha-value>)',
          hover: 'rgb(var(--color-surface-hover) / <alpha-value>)',
        },

        // Brand
        brand: {
          DEFAULT: 'rgb(var(--color-brand) / <alpha-value>)',
          hover: 'rgb(var(--color-brand-hover) / <alpha-value>)',
        },

        // Interaction
        upvote: 'rgb(var(--color-upvote) / <alpha-value>)',
        downvote: 'rgb(var(--color-downvote) / <alpha-value>)',
        cta: {
          DEFAULT: 'rgb(var(--color-cta) / <alpha-value>)',
          hover: 'rgb(var(--color-cta-hover) / <alpha-value>)',
        },

        // Thread depth colors
        thread: {
          0: 'rgb(var(--color-thread-0) / <alpha-value>)',
          1: 'rgb(var(--color-thread-1) / <alpha-value>)',
          2: 'rgb(var(--color-thread-2) / <alpha-value>)',
          3: 'rgb(var(--color-thread-3) / <alpha-value>)',
          4: 'rgb(var(--color-thread-4) / <alpha-value>)',
        },

        // === Existing palettes (updated primary scale) ===

        // Primary: Brand violet (shifted from Twitch #9146FF to own #7C3AED)
        primary: {
          50: '#f5f3ff',
          100: '#ede9fe',
          200: '#ddd6fe',
          300: '#c4b5fd',
          400: '#a78bfa',
          500: '#7C3AED',
          600: '#6D28D9',
          700: '#5B21B6',
          800: '#4C1D95',
          900: '#3B1578',
          950: '#2E1065',
        },

        // Secondary: neon-magenta vibe
        secondary: {
          50: '#fff0fb',
          100: '#ffe0f7',
          200: '#ffc2ef',
          300: '#ff94e2',
          400: '#ff5ed3',
          500: '#ff2bc2',
          600: '#e600a8',
          700: '#b80085',
          800: '#8f0067',
          900: '#6f0052',
          950: '#3f002e',
        },

        // Success state
        success: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
          950: '#052e16',
        },

        // Warning state
        warning: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
          950: '#451a03',
        },

        // Error state
        error: {
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
          950: '#450a0a',
        },

        // Info state (indigo — complements the violet brand)
        info: {
          50: '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          300: '#a5b4fc',
          400: '#818cf8',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
          950: '#1e1b4b',
        },

        // Neutral grays: purple-tinted for dark UI
        neutral: {
          50: '#f7f6fb',
          100: '#edeaf6',
          200: '#d6d0ea',
          300: '#b7aed6',
          400: '#9487bd',
          500: '#786aa3',
          600: '#5f5482',
          700: '#483f62',
          800: '#332b47',
          900: '#1f172d',
          950: '#0c0714',
        },
      },

      fontFamily: {
        sans: [
          'Inter',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          '"Segoe UI"',
          'Roboto',
          '"Helvetica Neue"',
          'Arial',
          'sans-serif',
        ],
        heading: [
          '"Space Grotesk"',
          'system-ui',
          '-apple-system',
          'sans-serif',
        ],
        accent: [
          'Syne',
          '"Space Grotesk"',
          'system-ui',
          'sans-serif',
        ],
        mono: [
          '"JetBrains Mono"',
          'ui-monospace',
          'SFMono-Regular',
          '"SF Mono"',
          'Menlo',
          'Consolas',
          'monospace',
        ],
      },

      zIndex: {
        dropdown: '1000',
        sticky: '1020',
        fixed: '1030',
        'modal-backdrop': '1040',
        modal: '1050',
        popover: '1060',
        tooltip: '1070',
      },

      keyframes: {
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'fade-out': {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-in-left': {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-in-up': {
          '0%': { transform: 'translateY(100%)' },
          '100%': { transform: 'translateY(0)' },
        },
        'slide-in-down': {
          '0%': { transform: 'translateY(-100%)' },
          '100%': { transform: 'translateY(0)' },
        },
        shimmer: {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(100%)' },
        },
        'vote-pulse': {
          '0%': { transform: 'scale(1)' },
          '50%': { transform: 'scale(1.2)' },
          '100%': { transform: 'scale(1)' },
        },
      },

      animation: {
        'fade-in': 'fade-in 0.2s ease-in-out',
        'fade-out': 'fade-out 0.2s ease-in-out',
        'slide-in-right': 'slide-in-right 0.3s ease-out',
        'slide-in-left': 'slide-in-left 0.3s ease-out',
        'slide-in-up': 'slide-in-up 0.3s ease-out',
        'slide-in-down': 'slide-in-down 0.3s ease-out',
        shimmer: 'shimmer 2s infinite',
        'vote-pulse': 'vote-pulse 200ms ease-out',
      },
    },
  },

  safelist: [
    // Primary button colors
    'bg-primary-500',
    'bg-primary-600',
    'bg-primary-700',
    'hover:bg-primary-600',
    'hover:bg-primary-700',
    'active:bg-primary-700',

    // Brand colors
    'bg-brand',
    'bg-brand-hover',
    'text-brand',
    'border-brand',

    // Secondary button colors
    'bg-secondary-500',
    'bg-secondary-600',
    'bg-secondary-700',
    'hover:bg-secondary-600',
    'hover:bg-secondary-700',
    'active:bg-secondary-700',

    // Error/danger colors
    'bg-error-500',
    'bg-error-600',
    'bg-error-700',
    'hover:bg-error-600',
    'hover:bg-error-700',
    'active:bg-error-700',

    // Surface colors
    'bg-surface',
    'bg-surface-raised',
    'bg-surface-hover',
    'hover:bg-surface-hover',

    // Text colors
    'text-white',
    'text-primary-500',
    'text-foreground',
    'text-text-primary',
    'text-text-secondary',
    'text-text-tertiary',
    'text-upvote',
    'text-downvote',
    'text-cta',

    // Border colors
    'border-primary-500',
    'border-2',
    'border-subtle',

    // Thread colors
    'border-thread-0',
    'border-thread-1',
    'border-thread-2',
    'border-thread-3',
    'border-thread-4',
  ],

  plugins: [],
};

export default config;
