import { useState, useEffect } from 'react';

export interface ScrollToTopProps {
  /** Threshold in pixels before button appears */
  threshold?: number;
  /** Additional CSS classes */
  className?: string;
}

export function ScrollToTop({ threshold = 500, className = '' }: ScrollToTopProps) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setIsVisible(window.scrollY > threshold);
    };

    window.addEventListener('scroll', handleScroll);
    handleScroll(); // Check initial state

    return () => window.removeEventListener('scroll', handleScroll);
  }, [threshold]);

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  if (!isVisible) return null;

  return (
    <button
      onClick={scrollToTop}
      className={`
        fixed bottom-24 right-4 xs:bottom-28 xs:right-8
        w-12 h-12 xs:w-14 xs:h-14 
        bg-primary-500 hover:bg-primary-600 
        dark:bg-primary-600 dark:hover:bg-primary-500
        text-white 
        rounded-full shadow-lg hover:shadow-xl
        transition-all duration-200 ease-in-out
        flex items-center justify-center 
        z-50 touch-target cursor-pointer
        group
        ${className}
      `}
      aria-label="Scroll to top"
      title="Scroll to top"
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
          d="M5 10l7-7m0 0l7 7m-7-7v18" 
        />
      </svg>
    </button>
  );
}
