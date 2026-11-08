import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Trophy } from 'lucide-react';
import { countdownApi, countdownPath, formatCountdownDate } from '@/lib/countdown-api';

interface CountdownFeaturesProps {
  clipId?: string;
  broadcasterId?: string;
}

/**
 * "Featured in countdown" marker for a clip or a broadcaster. Renders nothing
 * until a placement exists, and nothing when the lookup fails.
 */
export function CountdownFeatures({ clipId, broadcasterId }: CountdownFeaturesProps) {
  const params = clipId ? { clip_id: clipId } : broadcasterId ? { broadcaster_id: broadcasterId } : null;
  const { data } = useQuery({
    queryKey: ['countdown-features', clipId ?? null, broadcasterId ?? null],
    queryFn: () => countdownApi.featured(params!),
    enabled: params !== null,
    staleTime: 5 * 60 * 1000,
  });
  if (!data || data.length === 0) return null;

  const shown = data.slice(0, 3);
  return (
    <div className='mb-4 flex flex-wrap items-center gap-2' data-testid='countdown-features'>
      <span className='kicker inline-flex items-center gap-1.5'>
        <Trophy size={14} aria-hidden='true' /> Featured in countdown
      </span>
      {shown.map(feature => (
        <Link
          key={`${feature.edition_date}-${feature.series_slug}-${feature.clip_id}`}
          to={countdownPath(feature)}
          className='inline-flex min-h-9 items-center border border-line-strong px-3 text-sm hover:border-brand hover:text-foreground transition-colors'
        >
          #{feature.position} · {feature.series_name} · {formatCountdownDate(feature.edition_date, 'short')}
        </Link>
      ))}
      {data.length > shown.length && (
        <span className='text-sm text-muted-foreground'>and {data.length - shown.length} more</span>
      )}
    </div>
  );
}
