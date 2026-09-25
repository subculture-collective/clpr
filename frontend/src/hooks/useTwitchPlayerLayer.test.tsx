import { render, screen } from '@testing-library/react';
import { useRef } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  useOverlapsTwitchPlayer,
  useRegisterTwitchPlayer,
  useTwitchPlayerMounted,
} from './useTwitchPlayerLayer';

type Box = { left: number; top: number; width: number; height: number };

// Positions come from data-box="left,top,width,height" so each element has its own rect.
function stubRects() {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    const [left, top, width, height] = (this.dataset.box ?? '0,0,0,0').split(',').map(Number);
    return { left, top, width, height, right: left + width, bottom: top + height, x: left, y: top } as DOMRect;
  });
}

const box = ({ left, top, width, height }: Box) => `${left},${top},${width},${height}`;
const PLAYER = { left: 0, top: 100, width: 800, height: 450 };

function Player({ active = true }: { active?: boolean }) {
  const ref = useRef<HTMLDivElement>(null);
  useRegisterTwitchPlayer(ref, active);
  return <div ref={ref} data-box={box(PLAYER)} />;
}

function Floating({ at, children }: { at: Box; children?: React.ReactNode }) {
  const [covers, ref] = useOverlapsTwitchPlayer();
  const mounted = useTwitchPlayerMounted();
  return (
    <div ref={ref} data-box={box(at)} data-testid='floating' data-covers={String(covers)} data-mounted={String(mounted)}>
      {children}
    </div>
  );
}

describe('useTwitchPlayerLayer', () => {
  beforeEach(stubRects);
  afterEach(() => vi.restoreAllMocks());

  it('reports whether a player is mounted', () => {
    const { rerender, unmount } = render(<><Floating at={{ left: 900, top: 0, width: 40, height: 40 }} /><Player /></>);
    expect(screen.getByTestId('floating')).toHaveAttribute('data-mounted', 'true');
    rerender(<><Floating at={{ left: 900, top: 0, width: 40, height: 40 }} /><Player active={false} /></>);
    expect(screen.getByTestId('floating')).toHaveAttribute('data-mounted', 'false');
    unmount();
  });

  it('flags a floating element that sits on a player', () => {
    render(<><Player /><Floating at={{ left: 32, top: 500, width: 56, height: 56 }} /></>);
    expect(screen.getByTestId('floating')).toHaveAttribute('data-covers', 'true');
  });

  it('keeps clearance around the player edge', () => {
    render(<><Player /><Floating at={{ left: 804, top: 200, width: 40, height: 40 }} /></>);
    expect(screen.getByTestId('floating')).toHaveAttribute('data-covers', 'true');
  });

  it('ignores elements clear of the player', () => {
    render(<><Player /><Floating at={{ left: 900, top: 600, width: 56, height: 56 }} /></>);
    expect(screen.getByTestId('floating')).toHaveAttribute('data-covers', 'false');
  });

  it('does not treat its own miniplayer as covered', () => {
    render(
      <Floating at={{ left: 0, top: 100, width: 800, height: 600 }}>
        <Player />
      </Floating>,
    );
    expect(screen.getByTestId('floating')).toHaveAttribute('data-covers', 'false');
  });
});
