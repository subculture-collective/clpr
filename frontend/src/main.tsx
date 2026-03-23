import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { PlaybackProvider } from './context/PlaybackContext'
import { Buffer } from 'buffer'
import './index.css'
import './i18n' // Initialize i18n
import App from './App.tsx'
import { initSentry } from './lib/sentry'
import ErrorBoundary from './components/ErrorBoundary'
import { registerServiceWorker } from './lib/sw-register'

// Polyfill Buffer for gray-matter
globalThis.Buffer = Buffer;

// Force dark mode - add 'dark' class to root element
document.documentElement.classList.add('dark');

// Add global error handler to catch any module errors
window.addEventListener('error', (event) => {
  console.error('Global error caught:', {
    message: event.message,
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    error: event.error?.stack,
  })
})

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason)
})

// Initialize Sentry before rendering the app
try {
  initSentry({
    dsn: import.meta.env.VITE_SENTRY_DSN || '',
    environment: import.meta.env.VITE_SENTRY_ENVIRONMENT || import.meta.env.MODE,
    release: import.meta.env.VITE_SENTRY_RELEASE || '',
    tracesSampleRate: parseFloat(import.meta.env.VITE_SENTRY_TRACES_SAMPLE_RATE || '1.0'),
    enabled: import.meta.env.VITE_SENTRY_ENABLED === 'true',
  })
} catch (error) {
  console.error('Failed to initialize Sentry:', error)
}

// Create React Query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 1000 * 60 * 5, // 5 minutes
    },
  },
})

try {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ErrorBoundary>
        <QueryClientProvider client={queryClient}>
          <PlaybackProvider>
            <App />
          </PlaybackProvider>
        </QueryClientProvider>
      </ErrorBoundary>
    </StrictMode>,
  )
} catch (error) {
  console.error('Failed to render app:', error)
  const root = document.getElementById('root')
  if (root) {
    root.innerHTML = '<div style="padding: 20px; color: red;"><h1>Failed to load application</h1><p>Check the console for details.</p></div>'
  }
}

// Register service worker for PWA functionality
try {
  registerServiceWorker()
} catch (error) {
  console.error('Failed to register service worker:', error)
}
