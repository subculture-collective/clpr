import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { registerServiceWorker, unregisterServiceWorker, isPWAInstalled, canInstallPWA } from './sw-register';

describe('sw-register', () => {
  const originalNavigator = globalThis.navigator;
  const originalWindow = globalThis.window;

  beforeEach(() => {
    // Reset all mocks before each test
    vi.clearAllMocks();
    vi.unstubAllEnvs();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    // Restore original navigator and window
    Object.defineProperty(globalThis, 'navigator', {
      value: originalNavigator,
      writable: true,
      configurable: true,
    });
    Object.defineProperty(globalThis, 'window', {
      value: originalWindow,
      writable: true,
      configurable: true,
    });
  });

  describe('registerServiceWorker', () => {
    it('should skip registration in development mode', async () => {
      // In test/dev mode, it should return null
      const result = await registerServiceWorker();
      expect(result).toBeNull();
    });

    function productionRegistration(controlled = false) {
      vi.stubEnv('DEV', false);
      const worker = Object.assign(new EventTarget(), {
        state: 'installing',
        postMessage: vi.fn(),
      });
      const registration = Object.assign(new EventTarget(), {
        installing: worker,
        waiting: null as typeof worker | null,
      });
      const container = Object.assign(new EventTarget(), {
        controller: controlled ? worker : null,
        register: vi.fn().mockResolvedValue(registration),
      });
      const reload = vi.fn();
      const confirm = vi.fn().mockReturnValue(false);
      Object.defineProperty(globalThis, 'navigator', {
        value: { serviceWorker: container }, configurable: true,
      });
      Object.defineProperty(globalThis, 'window', {
        value: { location: { reload }, confirm }, configurable: true,
      });
      return { worker, registration, container, reload, confirm };
    }

    it('preserves the page when its first worker claims control', async () => {
      const { container, worker, reload, confirm } = productionRegistration();
      await registerServiceWorker();
      container.controller = worker;
      container.dispatchEvent(new Event('controllerchange'));
      expect(reload).not.toHaveBeenCalled();
      expect(confirm).not.toHaveBeenCalled();
    });

    it.each(['waiting', 'installing'] as const)(
      'activates an accepted %s update and reloads once after activation',
      async (state) => {
        const { registration, worker, container, reload, confirm } = productionRegistration(true);
        confirm.mockReturnValue(true);
        if (state === 'waiting') registration.waiting = worker;
        await registerServiceWorker();
        if (state === 'installing') {
          registration.dispatchEvent(new Event('updatefound'));
          worker.state = 'installed';
          worker.dispatchEvent(new Event('statechange'));
        }
        expect(worker.postMessage).toHaveBeenCalledExactlyOnceWith({ type: 'SKIP_WAITING' });
        expect(reload).not.toHaveBeenCalled();
        container.dispatchEvent(new Event('controllerchange'));
        container.dispatchEvent(new Event('controllerchange'));
        expect(reload).toHaveBeenCalledTimes(1);
      }
    );

    it('keeps the current page when an update is declined or another tab activates it', async () => {
      const { registration, worker, container, reload, confirm } = productionRegistration(true);
      registration.waiting = worker;
      await registerServiceWorker();
      expect(confirm).toHaveBeenCalledTimes(1);
      expect(worker.postMessage).not.toHaveBeenCalled();
      container.dispatchEvent(new Event('controllerchange'));
      expect(reload).not.toHaveBeenCalled();
    });

    it('should return null when service workers are not supported', async () => {
      // Mock environment as production
      vi.stubEnv('DEV', false);

      // Mock navigator without serviceWorker
      Object.defineProperty(globalThis, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      });

      const result = await registerServiceWorker();
      expect(result).toBeNull();

      vi.unstubAllEnvs();
    });
  });

  describe('unregisterServiceWorker', () => {
    it('should return false when service workers are not supported', async () => {
      // Mock navigator without serviceWorker
      Object.defineProperty(globalThis, 'navigator', {
        value: {},
        writable: true,
        configurable: true,
      });

      const result = await unregisterServiceWorker();
      expect(result).toBe(false);
    });
  });

  describe('isPWAInstalled', () => {
    it('should return false when not in standalone mode', () => {
      // Mock window.matchMedia
      Object.defineProperty(globalThis.window, 'matchMedia', {
        value: vi.fn().mockImplementation((query: string) => ({
          matches: false,
          media: query,
          onchange: null,
          addListener: vi.fn(),
          removeListener: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          dispatchEvent: vi.fn(),
        })),
        writable: true,
        configurable: true,
      });

      const result = isPWAInstalled();
      expect(result).toBe(false);
    });

    it('should return true when in standalone mode', () => {
      // Mock window.matchMedia to return standalone mode
      Object.defineProperty(globalThis.window, 'matchMedia', {
        value: vi.fn().mockImplementation((query: string) => ({
          matches: query === '(display-mode: standalone)',
          media: query,
          onchange: null,
          addListener: vi.fn(),
          removeListener: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          dispatchEvent: vi.fn(),
        })),
        writable: true,
        configurable: true,
      });

      const result = isPWAInstalled();
      expect(result).toBe(true);
    });
  });

  describe('canInstallPWA', () => {
    it('should return false when BeforeInstallPromptEvent is not available', () => {
      const result = canInstallPWA();
      expect(result).toBe(false);
    });

    it('should return true when BeforeInstallPromptEvent is available', () => {
      // Mock BeforeInstallPromptEvent
      Object.defineProperty(globalThis.window, 'BeforeInstallPromptEvent', {
        value: class {},
        writable: true,
        configurable: true,
      });

      const result = canInstallPWA();
      expect(result).toBe(true);
    });
  });
});
