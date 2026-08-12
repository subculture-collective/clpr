import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll, beforeEach, vi } from 'vitest';
import type { MockInstance } from 'vitest';
import { server } from './mocks/server';
import 'fake-indexeddb/auto';

type ConsoleLevel = 'error' | 'warn';
type ConsoleAllowance = {
    level: ConsoleLevel;
    pattern: RegExp;
    remaining: number;
};

let consoleAllowances: ConsoleAllowance[] = [];
let capturedConsole: Array<{ level: ConsoleLevel; message: string }> = [];
let consoleErrorSpy: MockInstance | undefined;
let consoleWarnSpy: MockInstance | undefined;

const formatConsoleArguments = (args: unknown[]): string =>
    args
        .map(value => {
            if (value instanceof Error) return value.stack || value.message;
            if (typeof value === 'string') return value;
            try {
                return JSON.stringify(value);
            } catch {
                return String(value);
            }
        })
        .join(' ');

/**
 * Declare an intentional warning/error in the owning test. The allowance must
 * be consumed exactly the requested number of times or the test fails.
 */
export function allowTestConsole(
    level: ConsoleLevel,
    pattern: RegExp,
    count = 1,
): void {
    consoleAllowances.push({ level, pattern, remaining: count });
}

beforeEach(() => {
    consoleAllowances = [];
    capturedConsole = [];
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation((...args) => {
        capturedConsole.push({ level: 'error', message: formatConsoleArguments(args) });
    });
    consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation((...args) => {
        capturedConsole.push({ level: 'warn', message: formatConsoleArguments(args) });
    });
});

afterEach(() => {
    consoleErrorSpy?.mockRestore();
    consoleWarnSpy?.mockRestore();

    const unexpected: typeof capturedConsole = [];
    for (const entry of capturedConsole) {
        const allowance = consoleAllowances.find(
            candidate =>
                candidate.level === entry.level &&
                candidate.remaining > 0 &&
                candidate.pattern.test(entry.message),
        );
        if (allowance) {
            allowance.remaining -= 1;
        } else {
            unexpected.push(entry);
        }
    }

    const unused = consoleAllowances.filter(allowance => allowance.remaining > 0);
    if (unexpected.length || unused.length) {
        const details = [
            ...unexpected.map(entry => `unexpected console.${entry.level}: ${entry.message}`),
            ...unused.map(
                allowance =>
                    `unused console.${allowance.level} allowance ${allowance.pattern} x${allowance.remaining}`,
            ),
        ];
        throw new Error(`Test console contract failed:\n${details.join('\n')}`);
    }
});

// Full-document navigation is intentionally not emulated by jsdom. Components
// with redirect behavior expose injectable boundaries in unit tests; actual
// browser navigation belongs to the Playwright smoke suite.

// Node.js 22+ introduces built-in localStorage/sessionStorage globals that require
// --localstorage-file to function. These broken globals shadow jsdom's working
// implementations. Provide a proper in-memory Storage polyfill.
function createStorage(): Storage {
    let store: Record<string, string> = {};
    return {
        getItem(key: string): string | null {
            return key in store ? store[key] : null;
        },
        setItem(key: string, value: string): void {
            store[key] = String(value);
        },
        removeItem(key: string): void {
            delete store[key];
        },
        clear(): void {
            store = {};
        },
        key(index: number): string | null {
            const keys = Object.keys(store);
            return keys[index] ?? null;
        },
        get length(): number {
            return Object.keys(store).length;
        },
    };
}

// Always replace the globals. Node's built-in Storage methods exist and pass a
// typeof check even when they throw because no --localstorage-file was set.
const ls = createStorage();
Object.defineProperty(globalThis, 'localStorage', { value: ls, writable: true, configurable: true });
Object.defineProperty(window, 'localStorage', { value: ls, writable: true, configurable: true });

const ss = createStorage();
Object.defineProperty(globalThis, 'sessionStorage', { value: ss, writable: true, configurable: true });
Object.defineProperty(window, 'sessionStorage', { value: ss, writable: true, configurable: true });

// Unhandled application requests indicate incomplete test setup and must fail.
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

// Reset any request handlers that we may add during the tests,
// so they don't affect other tests
afterEach(() => server.resetHandlers());

// Clean up after all tests are done
afterAll(() => server.close());

// Cleanup after each test
afterEach(() => {
    cleanup();
});

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: () => {}, // deprecated
        removeListener: () => {}, // deprecated
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => {},
    }),
});

// Mock IntersectionObserver
class MockIntersectionObserver {
    readonly root: Element | Document | null = null;
    readonly rootMargin: string = '';
    readonly thresholds: ReadonlyArray<number> = [];
    constructor() {}
    observe(): void {}
    unobserve(): void {}
    disconnect(): void {}
    takeRecords(): IntersectionObserverEntry[] {
        return [];
    }
}

globalThis.IntersectionObserver =
    MockIntersectionObserver as unknown as typeof globalThis.IntersectionObserver;

// Mock ResizeObserver for chart components
class MockResizeObserver {
    observe(): void {}
    unobserve(): void {}
    disconnect(): void {}
}

globalThis.ResizeObserver =
    MockResizeObserver as unknown as typeof globalThis.ResizeObserver;

Object.defineProperty(window, 'scrollTo', {
    configurable: true,
    writable: true,
    value: vi.fn(),
});

const nativeGetComputedStyle = window.getComputedStyle.bind(window);
Object.defineProperty(window, 'getComputedStyle', {
    configurable: true,
    writable: true,
    // jsdom does not implement pseudo-element styles. Axe only needs the base
    // computed style for these component tests.
    value: (element: Element) => nativeGetComputedStyle(element),
});

// jsdom intentionally omits canvas rendering. Tests only need a stable context
// boundary for libraries that perform feature detection.
Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
    configurable: true,
    value: () => ({
        getImageData: () => ({ data: new Uint8ClampedArray(4) }),
        measureText: () => ({ width: 0 }),
        save: () => {},
        restore: () => {},
        scale: () => {},
        clearRect: () => {},
        fillRect: () => {},
    }),
});
