import { useLayoutEffect, useState, useSyncExternalStore, type RefObject } from 'react';

// Twitch forbids page elements from covering an embed. Players register the box
// that holds their iframe; floating controls and the consent banner read the
// registry to stay clear of it.
// https://dev.twitch.tv/docs/embed/

const players = new Set<HTMLElement>();
const listeners = new Set<() => void>();

function notify() {
  listeners.forEach((listener) => listener());
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

const hasPlayer = () => players.size > 0;

/** Register the element holding a Twitch iframe while `active` is true. */
export function useRegisterTwitchPlayer(ref: RefObject<HTMLElement | null>, active: boolean) {
  useLayoutEffect(() => {
    const element = ref.current;
    if (!active || !element) return;
    players.add(element);
    notify();
    return () => {
      players.delete(element);
      notify();
    };
  }, [ref, active]);
}

/** Whether any Twitch player is currently mounted. */
export function useTwitchPlayerMounted(): boolean {
  return useSyncExternalStore(subscribe, hasPlayer, () => false);
}

// Floating elements keep this much clearance from a player.
const CLEARANCE = 8;
// Players can move without a scroll or resize (content loading above them).
const RECHECK_MS = 500;

function overlapsPlayer(element: HTMLElement): boolean {
  const rect = element.getBoundingClientRect();
  if (rect.width === 0 && rect.height === 0) return false;
  for (const player of players) {
    // A player inside the floating element (the queue miniplayer) is its own content.
    if (element.contains(player)) continue;
    const box = player.getBoundingClientRect();
    if (
      rect.left < box.right + CLEARANCE &&
      rect.right > box.left - CLEARANCE &&
      rect.top < box.bottom + CLEARANCE &&
      rect.bottom > box.top - CLEARANCE
    ) {
      return true;
    }
  }
  return false;
}

/**
 * Whether a fixed/floating element would cover a registered Twitch player.
 * Attach the returned callback ref to the element; callers hide it with
 * `visibility: hidden` (which keeps it measurable) while this is true.
 */
export function useOverlapsTwitchPlayer(
  enabled = true,
): [boolean, (element: HTMLElement | null) => void] {
  const mounted = useTwitchPlayerMounted();
  const [element, setElement] = useState<HTMLElement | null>(null);
  const [overlapping, setOverlapping] = useState(false);
  const watching = enabled && mounted && element !== null;

  useLayoutEffect(() => {
    // The returned value is gated on `watching`, so stale state is harmless.
    if (!watching || !element) return;
    let frame = 0;
    const measure = () => setOverlapping(overlapsPlayer(element));
    const schedule = () => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(measure);
    };
    measure();
    const unsubscribe = subscribe(schedule);
    const interval = window.setInterval(measure, RECHECK_MS);
    window.addEventListener('scroll', schedule, { capture: true, passive: true });
    window.addEventListener('resize', schedule);
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(schedule);
    observer?.observe(element);
    players.forEach((player) => observer?.observe(player));
    return () => {
      cancelAnimationFrame(frame);
      unsubscribe();
      window.clearInterval(interval);
      window.removeEventListener('scroll', schedule, { capture: true });
      window.removeEventListener('resize', schedule);
      observer?.disconnect();
    };
  }, [element, watching]);

  return [watching && overlapping, setElement];
}
