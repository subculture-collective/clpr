import { useQuery } from '@tanstack/react-query';
import { Download } from 'lucide-react';
import { Card, CardBody, CardHeader } from '@/components';
import { clipPackApi, type ClipPack } from '@/lib/clip-pack-api';
import { formatCountdownDate } from '@/lib/countdown-api';

function megabytes(bytes?: number): string {
  return bytes ? `${(bytes / 1_000_000).toFixed(bytes < 10_000_000 ? 1 : 0)} MB` : '';
}

function PackRow({ pack }: { pack: ClipPack }) {
  const week = `Week of ${formatCountdownDate(pack.week_start, 'short')}`;
  let detail: string;
  switch (pack.withdrawn && pack.status === 'ready' ? 'withdrawn' : pack.status) {
    case 'withdrawn':
      detail = 'No longer available: a clip in it was removed, or you stopped featuring your clips';
      break;
    case 'ready':
      detail = [
        `${pack.clip_count ?? 0} clips`,
        megabytes(pack.bytes),
        pack.expires_at && `available until ${new Date(pack.expires_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}`,
      ]
        .filter(Boolean)
        .join(' · ');
      break;
    case 'pending':
    case 'building':
      detail = 'Being made…';
      break;
    case 'expired':
      detail = 'Expired after 7 days';
      break;
    default:
      detail = 'Could not be made: none of the clips could be downloaded';
  }
  return (
    <li className='flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0'>
      <div>
        <p className='font-semibold text-foreground'>{week}</p>
        <p className='text-sm text-muted-foreground'>{detail}</p>
      </div>
      {pack.status === 'ready' && !pack.withdrawn && pack.download_url && (
        <a
          href={pack.download_url}
          download
          className='inline-flex min-h-11 items-center gap-2 bg-brand px-4 text-sm font-semibold text-background hover:opacity-90'
        >
          <Download size={16} aria-hidden='true' /> Download<span className='sr-only'> the pack for the {week.toLowerCase()}</span>
        </a>
      )}
    </li>
  );
}

/**
 * The creator dashboard's weekly clip packs (#86): each Monday, a zip of the
 * broadcaster's top clips of the past week as captioned 9:16 files and a
 * 16:9 highlight reel. Shown once the broadcaster has a pack.
 */
export function ClipPacksCard() {
  const { data } = useQuery({ queryKey: ['clip-packs'], queryFn: clipPackApi.mine, retry: false, staleTime: 60 * 1000 });
  if (!data || data.length === 0) return null;
  return (
    <Card className='mb-8' data-testid='clip-packs'>
      <CardHeader>
        <h2 className='text-xl font-semibold'>Your weekly clip packs</h2>
      </CardHeader>
      <CardBody>
        <p className='mb-4 text-sm text-muted-foreground'>
          Every Monday, your most-viewed clips of the past week as captioned vertical videos and a highlight reel, ready to
          post. Each pack is kept for 7 days.
        </p>
        <ul className='divide-y divide-border' aria-label='Your clip packs'>
          {data.map(pack => (
            <PackRow key={pack.id} pack={pack} />
          ))}
        </ul>
      </CardBody>
    </Card>
  );
}
