import { Link } from 'react-router-dom';
import { Button } from '@/components/ui';
import { useAuth } from '@/context/AuthContext';

/**
 * The countdown page's offer to streamers (#84): connect the channel, get
 * featured in the daily countdowns, and see where each clip placed.
 */
export function CreatorConnectOffer() {
  const { user } = useAuth();
  const dashboard = user?.twitch_id ? `/creator/${encodeURIComponent(user.twitch_id)}/dashboard` : null;
  return (
    <section className='mb-10 border border-line-strong p-6' aria-labelledby='creator-connect-title'>
      <p className='kicker mb-2'>For streamers</p>
      <h2 id='creator-connect-title' className='text-2xl font-bold text-foreground mb-3'>
        Get your clips into the daily countdown
      </h2>
      <ul className='mb-5 space-y-1.5 text-muted-foreground'>
        <li>Connect your Twitch channel so clpr can feature your clips, credited to you.</li>
        <li>Rank in the daily countdowns and the Shorts made from them on YouTube.</li>
        <li>See on your creator dashboard where each clip placed and the views it got.</li>
      </ul>
      {dashboard ? (
        <Button asChild variant='primary'>
          <Link to={dashboard}>Open my creator dashboard</Link>
        </Button>
      ) : (
        <Button asChild variant='primary'>
          <Link to='/login' state={{ from: { pathname: '/countdowns' } }}>
            Sign in with Twitch
          </Link>
        </Button>
      )}
    </section>
  );
}
