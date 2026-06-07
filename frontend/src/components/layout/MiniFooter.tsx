import { useState } from 'react';
import { Link } from 'react-router-dom';

export function MiniFooter() {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="fixed bottom-4 left-4 xs:bottom-8 xs:left-8 z-40">
      {/* Collapsed state - Icon button */}
      {!isExpanded && (
        <button
          onClick={() => setIsExpanded(true)}
          className="
            w-12 h-12 xs:w-14 xs:h-14
            bg-card border-2 border-border
            hover:bg-muted
            text-foreground
            rounded-full shadow-lg hover:shadow-xl
            transition-all duration-200 ease-in-out
            flex items-center justify-center
            touch-target
            group
          "
          aria-label="Show footer links"
          title="Footer links"
        >
          <svg
            className="w-6 h-6 transition-transform duration-200 group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </button>
      )}

      {/* Expanded state - Footer links */}
      {isExpanded && (
        <div className="
          bg-card border-2 border-border
          rounded-2xl shadow-2xl
          p-4 xs:p-6
          max-w-xs xs:max-w-sm
          animate-slide-in-up
        ">
          {/* Header with close button */}
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-foreground">Quick Links</h3>
            <button
              onClick={() => setIsExpanded(false)}
              className="
                w-8 h-8
                hover:bg-muted
                rounded-full
                transition-colors
                flex items-center justify-center
              "
              aria-label="Close footer links"
            >
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          {/* Footer links */}
          <div className="space-y-3">
            <div>
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                About
              </h4>
              <div className="space-y-1">
                <Link
                  to="/about"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                  onClick={() => setIsExpanded(false)}
                >
                  About clpr
                </Link>
                <a
                  href="https://git.subcult.tv/subculture-collective/clpr"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                >
                  GitHub
                </a>
              </div>
            </div>

            <div>
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                Legal
              </h4>
              <div className="space-y-1">
                <Link
                  to="/privacy"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                  onClick={() => setIsExpanded(false)}
                >
                  Privacy Policy
                </Link>
                <Link
                  to="/terms"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                  onClick={() => setIsExpanded(false)}
                >
                  Terms of Service
                </Link>
                <Link
                  to="/legal/dmca"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                  onClick={() => setIsExpanded(false)}
                >
                  DMCA Policy
                </Link>
              </div>
            </div>

            <div>
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                Community
              </h4>
              <div className="space-y-1">
                <a
                  href="https://discord.gg/TFwB4aJRef"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                >
                  Discord
                </a>
                <a
                  href="https://x.com/clpr_tv"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block text-sm text-foreground hover:text-primary-500 transition-colors"
                >
                  Twitter
                </a>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
