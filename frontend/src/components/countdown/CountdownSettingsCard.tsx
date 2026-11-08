import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Card, CardBody, CardHeader, Toggle } from '@/components';
import { countdownControlsApi, featureMyClipsUrl, type CountdownControls } from '@/lib/countdown-controls-api';
import { getErrorMessage, getHttpStatus } from '@/lib/error-utils';

/** The signed-in broadcaster's countdown controls (#75). */
export function CountdownSettingsCard() {
  const queryClient = useQueryClient();
  const { data, error, isLoading } = useQuery({
    queryKey: ['countdown-controls'],
    queryFn: countdownControlsApi.get,
    retry: false,
  });
  const update = (next: CountdownControls) => queryClient.setQueryData(['countdown-controls'], next);
  const exclusion = useMutation({ mutationFn: countdownControlsApi.setExcluded, onSuccess: update });
  const stopFeaturing = useMutation({ mutationFn: countdownControlsApi.stopFeaturing, onSuccess: update });

  let body;
  if (isLoading) {
    body = <p className='text-muted-foreground'>Loading…</p>;
  } else if (getHttpStatus(error) === 409) {
    body = <p className='text-muted-foreground'>Sign in with Twitch to manage how your clips appear in countdowns.</p>;
  } else if (error || !data) {
    body = <Alert variant='error'>Your countdown settings could not be loaded.</Alert>;
  } else {
    const byStaff = data.excluded && data.exclusion_source === 'admin';
    const mutationError = exclusion.error ?? stopFeaturing.error;
    body = (
      <div className='space-y-6'>
        <p className='text-sm text-muted-foreground'>
          clpr&apos;s daily countdowns and Shorts on YouTube use public Twitch clips, each credited to its broadcaster.
        </p>
        <Toggle
          label='Leave my clips out'
          helperText={
            byStaff
              ? 'Your clips are left out at your request. Contact us to change this.'
              : 'Keeps your clips out of new countdowns and off the countdown pages. Clips already in YouTube videos are removed, or the videos edited, within 7 days.'
          }
          checked={data.excluded}
          disabled={byStaff || exclusion.isPending}
          onChange={event => exclusion.mutate(event.target.checked)}
        />
        <div className='border-t border-border pt-5'>
          <h3 className='font-semibold text-foreground mb-1'>Feature my clips</h3>
          {data.featured ? (
            <>
              <p className='text-sm text-muted-foreground mb-3'>
                Your clips are featured: clpr downloads them through Twitch&apos;s official clip download API.
              </p>
              <Button variant='outline' disabled={stopFeaturing.isPending} onClick={() => stopFeaturing.mutate()}>
                Stop featuring my clips
              </Button>
            </>
          ) : (
            <>
              <p className='text-sm text-muted-foreground mb-3'>
                Let clpr download your clips through Twitch&apos;s official API, so they can be featured in countdowns and in
                clip packs made for you. Twitch asks you to allow &ldquo;Manage clips&rdquo;; you can disconnect at any time.
              </p>
              <Button asChild variant='primary'>
                <a href={featureMyClipsUrl()}>Feature my clips</a>
              </Button>
            </>
          )}
        </div>
        {mutationError && <Alert variant='error'>{getErrorMessage(mutationError, 'The change could not be saved.')}</Alert>}
        <p className='text-xs text-muted-foreground'>Wording version {data.consent_text_version}</p>
      </div>
    );
  }

  return (
    <Card className='mb-6' data-testid='countdown-settings'>
      <CardHeader>
        <h2 className='text-xl font-semibold'>Your clips in countdowns</h2>
      </CardHeader>
      <CardBody>{body}</CardBody>
    </Card>
  );
}
