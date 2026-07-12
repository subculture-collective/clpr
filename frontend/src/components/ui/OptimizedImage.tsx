import { useState, useEffect, useRef } from 'react';
import { cn } from '@/lib/utils';

export interface OptimizedImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  /**
   * Image source URL
   */
  src: string;
  /**
   * Alt text for the image
   */
  alt: string;
  /**
   * Aspect ratio to reserve space and prevent CLS (e.g., "16/9", "4/3", "1/1")
   */
  aspectRatio?: string;
  /**
   * Sizes attribute for responsive images
   */
  sizes?: string;
  /**
   * Priority loading - disables lazy loading for above-the-fold images
   */
  priority?: boolean;
  /**
   * Blur placeholder while loading
   */
  blurDataURL?: string;
}

/**
 * Optimized image component with lazy loading, aspect ratio preservation for CLS prevention,
 * and responsive sizing for Core Web Vitals improvements
 */
export function OptimizedImage({
  src,
  alt,
  aspectRatio,
  sizes,
  priority = false,
  blurDataURL,
  className,
  onLoad,
  onError,
  ...props
}: OptimizedImageProps) {
  const [isLoaded, setIsLoaded] = useState(false);
  const [hasError, setHasError] = useState(false);
  const imgRef = useRef<HTMLImageElement>(null);

  const intrinsicAspectRatio =
    typeof props.width === 'number' && typeof props.height === 'number'
      ? `${props.width} / ${props.height}`
      : undefined;

  useEffect(() => {
    // For priority images, start loading immediately
    if (priority && imgRef.current?.complete) {
      queueMicrotask(() => setIsLoaded(true));
    }
  }, [priority]);

  const handleLoad = (e: React.SyntheticEvent<HTMLImageElement>) => {
    setIsLoaded(true);
    onLoad?.(e);
  };

  const handleError = (e: React.SyntheticEvent<HTMLImageElement>) => {
    setHasError(true);
    onError?.(e);
  };

  // Container styles with aspect ratio to prevent layout shift
  const containerStyles = aspectRatio || intrinsicAspectRatio
    ? { aspectRatio: aspectRatio ?? intrinsicAspectRatio }
    : undefined;

  return (
    <div
      className={cn('relative overflow-hidden bg-neutral-100 dark:bg-neutral-800', className)}
      style={containerStyles}
    >
      {/* Blur placeholder */}
      {!isLoaded && blurDataURL && !hasError && (
        <img
          src={blurDataURL}
          alt=""
          aria-hidden="true"
          className="absolute inset-0 w-full h-full object-cover blur-sm"
        />
      )}

      {/* Main image */}
      {!hasError && (
        <img
          ref={imgRef}
          src={src}
          alt={alt}
          sizes={sizes}
          loading={priority ? 'eager' : 'lazy'}
          decoding="async"
          // Note: fetchPriority has limited browser support (not in Firefox as of early 2024)
          // See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/img#browser_compatibility
          fetchPriority={priority ? 'high' : 'auto'}
          onLoad={handleLoad}
          onError={handleError}
          className={cn(
            'w-full h-full object-cover transition-opacity duration-300 motion-reduce:transition-none',
            isLoaded ? 'opacity-100' : 'opacity-0'
          )}
          {...props}
        />
      )}

      {/* Error fallback */}
      {hasError && (
        <div className="absolute inset-0 flex items-center justify-center bg-neutral-100 dark:bg-neutral-800">
          <div className="text-center p-4">
            <svg
              className="w-12 h-12 mx-auto mb-2 text-neutral-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            <p className="text-sm text-neutral-500">Image unavailable</p>
          </div>
        </div>
      )}
    </div>
  );
}
