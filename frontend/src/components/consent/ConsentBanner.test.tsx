import { fireEvent, render, screen } from '@testing-library/react';
import { useRef } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useRegisterTwitchPlayer } from '@/hooks/useTwitchPlayerLayer';
import { CONSENT_BANNER_SLOT_ID, ConsentBanner } from './ConsentBanner';

const acceptAll = vi.fn();
vi.mock('../../context/ConsentContext', () => ({
  useConsent: () => ({
    showConsentBanner: true,
    consent: { functional: false, analytics: false, advertising: false },
    updateConsent: vi.fn(),
    acceptAll,
    rejectAll: vi.fn(),
    doNotTrack: false,
  }),
}));

function Player() {
  const ref = useRef<HTMLDivElement>(null);
  useRegisterTwitchPlayer(ref, true);
  return <div ref={ref} data-testid='player' />;
}

// 1366×680 viewport: the overlay banner occupies the bottom 120px.
function stubRects(playerBottom: number) {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    const rect = this.dataset.testid === 'player'
      ? { left: 100, top: 100, right: 1000, bottom: playerBottom, width: 900, height: playerBottom - 100 }
      : this.dataset.placement
        ? { left: 0, top: 560, right: 1366, bottom: 680, width: 1366, height: 120 }
        : { left: 0, top: 0, right: 0, bottom: 0, width: 0, height: 0 };
    return rect as DOMRect;
  });
}

function renderPage(withPlayer = true) {
  return render(
    <MemoryRouter>
      <div id={CONSENT_BANNER_SLOT_ID} data-testid='slot' />
      {withPlayer && <Player />}
      <ConsentBanner />
    </MemoryRouter>,
  );
}

describe('ConsentBanner placement around Twitch players', () => {
  beforeEach(() => acceptAll.mockReset());
  afterEach(() => {
    vi.restoreAllMocks();
    document.body.style.paddingBottom = '';
    document.documentElement.style.removeProperty('--consent-banner-height');
  });

  it('moves into the page flow instead of covering a player', () => {
    stubRects(640);
    renderPage();
    const banner = screen.getByRole('region', { name: 'Privacy & Cookie Preferences' });
    expect(banner).toHaveAttribute('data-placement', 'inline');
    expect(screen.getByTestId('slot')).toContainElement(banner);
    expect(banner).not.toHaveClass('fixed');
    expect(document.body.style.paddingBottom).toBe('');
    expect(document.documentElement.style.getPropertyValue('--consent-banner-height')).toBe('');

    // The consent flow still works from the in-page banner.
    fireEvent.click(screen.getByRole('button', { name: 'Accept All' }));
    expect(acceptAll).toHaveBeenCalledTimes(1);
  });

  it('stays a bottom overlay and reserves space when no player is underneath', () => {
    stubRects(500);
    renderPage();
    const banner = screen.getByRole('region', { name: 'Privacy & Cookie Preferences' });
    expect(banner).toHaveAttribute('data-placement', 'overlay');
    expect(banner).toHaveClass('fixed');
    expect(screen.getByTestId('slot')).not.toContainElement(banner);
    expect(document.documentElement.style.getPropertyValue('--consent-banner-height')).toBe('120px');
    expect(document.body.style.paddingBottom).toBe('120px');
  });

  it('stays a bottom overlay on pages without a player', () => {
    stubRects(640);
    renderPage(false);
    expect(screen.getByRole('region', { name: 'Privacy & Cookie Preferences' })).toHaveAttribute('data-placement', 'overlay');
  });
});
