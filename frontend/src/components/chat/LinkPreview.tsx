import { ExternalLink } from 'lucide-react';
import { useState, useEffect } from 'react';

interface LinkPreviewProps {
  url: string;
}

interface LinkMetadata {
  title?: string;
  description?: string;
  image?: string;
  url: string;
}

/**
 * LinkPreview component that fetches and displays metadata for URLs
 */
export function LinkPreview({ url }: LinkPreviewProps) {
  const [metadata, setMetadata] = useState<LinkMetadata | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    // In a production environment, this would call a backend API
    // that safely fetches and parses the URL metadata
    // For now, we'll just show a basic link preview
    const fetchMetadata = async () => {
      try {
        setLoading(true);
        setError(false);

        // Mock metadata fetch - in production, call your backend API
        // const response = await fetch(`/api/link-preview?url=${encodeURIComponent(url)}`);
        // const data = await response.json();
        
        // For now, just extract domain name
        const urlObj = new URL(url);
        setMetadata({
          title: urlObj.hostname,
          description: url,
          url: url,
        });
      } catch (err) {
        console.error('Error fetching link preview:', err);
        setError(true);
      } finally {
        setLoading(false);
      }
    };

    fetchMetadata();
  }, [url]);

  if (loading || error || !metadata) {
    return (
      <a
        href={url}
        target="_blank"
        rel="noopener noreferrer"
        className="text-link dark:text-primary-400 hover:underline break-all inline-flex items-center gap-1"
      >
        {url}
        <ExternalLink className="w-3 h-3 flex-shrink-0" />
      </a>
    );
  }

  return (
    <div className="my-2 border border-border rounded-lg overflow-hidden hover:border-primary-500 transition-colors">
      <a
        href={metadata.url}
        target="_blank"
        rel="noopener noreferrer"
        className="block no-underline"
      >
        {metadata.image && (
          <img
            src={metadata.image}
            alt={metadata.title || 'Link preview'}
            className="w-full h-32 object-cover"
            loading="lazy"
          />
        )}
        <div className="p-3">
          {metadata.title && (
            <div className="font-medium text-sm text-foreground mb-1 line-clamp-1">
              {metadata.title}
            </div>
          )}
          {metadata.description && (
            <div className="text-xs text-muted-foreground line-clamp-2">
              {metadata.description}
            </div>
          )}
          <div className="flex items-center gap-1 mt-2 text-xs text-link dark:text-primary-400">
            <ExternalLink className="w-3 h-3" />
            <span className="truncate">{new URL(metadata.url).hostname}</span>
          </div>
        </div>
      </a>
    </div>
  );
}
