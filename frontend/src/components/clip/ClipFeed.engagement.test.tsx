import { render, screen } from '@/test/test-utils';
import { describe, expect, it, vi } from 'vitest';
import * as clipApi from '@/lib/clip-api';
import { ClipFeed } from './ClipFeed';
import type { ClipFeedResponse, EngagementMetadata } from '@/types/clip';

vi.mock('@/lib/clip-api');

function feedWith(engagement: EngagementMetadata): ClipFeedResponse {
    return { clips: [], total: 0, page: 1, has_more: false, limit: 20, engagement };
}

describe('ClipFeed engagement note', () => {
    it('summarises partial coverage in one friendly line', async () => {
        // Production shape on first visit: partial coverage with some stale observations.
        vi.mocked(clipApi.fetchClips).mockResolvedValue(feedWith({
            period: 'day',
            partial_coverage: true,
            estimated: true,
            stale_clips: 753,
            eligible_clips: 41692,
        }));

        render(<ClipFeed />);

        expect(await screen.findByText('Ranked by new views, votes, and comments. Estimates cover what clpr has tracked so far.')).toBeInTheDocument();
        expect(screen.queryByText(/observations are stale/)).not.toBeInTheDocument();
        expect(screen.queryByText(/History is incomplete/)).not.toBeInTheDocument();
    });

    it('keeps the estimate disclosure when coverage is complete', async () => {
        vi.mocked(clipApi.fetchClips).mockResolvedValue(feedWith({
            period: 'week',
            partial_coverage: false,
            estimated: true,
        }));

        render(<ClipFeed />);

        expect(await screen.findByText('Ranked by new views, votes, and comments (estimated).')).toBeInTheDocument();
    });
});
