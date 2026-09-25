import { render, screen } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { TwitchChatEmbed } from './TwitchChatEmbed';

vi.mock('../../context/AuthContext', () => ({ useAuth: () => ({ isAuthenticated: false }) }));
vi.mock('../../lib/twitch-api', () => ({ checkTwitchAuthStatus: vi.fn() }));

it('frames Twitch chat from www.twitch.tv with this site as parent', () => {
  render(<TwitchChatEmbed channel='some streamer' />);
  const src = new URL(screen.getByTitle('some streamer Twitch Chat').getAttribute('src')!);
  // frame-src in every CSP copy must allow this origin.
  expect(src.origin).toBe('https://www.twitch.tv');
  expect(src.pathname).toBe('/embed/some%20streamer/chat');
  expect(src.searchParams.get('parent')).toBe(window.location.hostname);
});
