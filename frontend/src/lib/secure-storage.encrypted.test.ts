import { webcrypto } from 'node:crypto';
import { IDBFactory, IDBObjectStore } from 'fake-indexeddb';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { allowTestConsole } from '../test/setup';
import { clearSecureStorage, getSecureItem, removeSecureItem, setSecureItem } from './secure-storage';

beforeEach(() => {
    vi.stubGlobal('crypto', webcrypto);
    vi.stubGlobal('indexedDB', new IDBFactory());
    sessionStorage.clear();
    localStorage.clear();
});
afterEach(() => vi.unstubAllGlobals());

describe('encrypted session credentials', () => {
    it('round-trips Unicode with a session key and replaces old values', async () => {
        await setSecureItem('token', 'first');
        const key = sessionStorage.getItem('clpr-encryption-key');
        await setSecureItem('token', 'secret — 🔐');
        expect(await getSecureItem('token')).toBe('secret — 🔐');
        expect(sessionStorage.getItem('clpr-encryption-key')).toBe(key);
        expect(sessionStorage.getItem('secure_token')).toBeNull();
        expect(localStorage.length).toBe(0);
    });

    it('removes only the requested credential from both storage tiers', async () => {
        await setSecureItem('access', 'access-token');
        await setSecureItem('refresh', 'refresh-token');
        sessionStorage.setItem('secure_access', 'legacy');
        localStorage.setItem('secure_access', 'legacy');
        await removeSecureItem('access');
        expect(await getSecureItem('access')).toBeNull();
        expect(await getSecureItem('refresh')).toBe('refresh-token');
        expect(localStorage.getItem('secure_access')).toBeNull();
    });

    it('logout clears all encrypted and legacy credentials without removing preferences', async () => {
        await setSecureItem('access', 'access-token');
        await setSecureItem('refresh', 'refresh-token');
        sessionStorage.setItem('secure_legacy', 'legacy');
        localStorage.setItem('secure_legacy', 'legacy');
        localStorage.setItem('theme', 'dark');
        await clearSecureStorage();
        expect(await getSecureItem('access')).toBeNull();
        expect(await getSecureItem('refresh')).toBeNull();
        expect(sessionStorage.getItem('secure_legacy')).toBeNull();
        expect(localStorage.getItem('secure_legacy')).toBeNull();
        expect(localStorage.getItem('theme')).toBe('dark');
    });

    it('cannot recover old ciphertext after the session key is lost', async () => {
        await setSecureItem('token', 'sensitive');
        sessionStorage.clear();
        allowTestConsole('error', /Error retrieving secure item/);
        expect(await getSecureItem('token')).toBeNull();
    });

    it.each(['corrupt', '{}'])('replaces invalid key data %s when storing a new credential', async (key) => {
        sessionStorage.setItem('clpr-encryption-key', key);
        await setSecureItem('token', 'new-token');
        expect(await getSecureItem('token')).toBe('new-token');
    });

    it('falls back to session storage after an aborted write and removes stale fallback after recovery', async () => {
        await setSecureItem('token', 'old-value');
        const put = IDBObjectStore.prototype.put;
        const spy = vi.spyOn(IDBObjectStore.prototype, 'put').mockImplementationOnce(function (...args) {
            const request = put.apply(this, args);
            this.transaction.abort();
            return request;
        });
        allowTestConsole('error', /Error storing secure item/);
        await setSecureItem('token', 'fallback');
        spy.mockRestore();
        expect(await getSecureItem('token')).toBe('fallback');
        await setSecureItem('token', 'committed');
        expect(await getSecureItem('token')).toBe('committed');
        expect(sessionStorage.getItem('secure_token')).toBeNull();
    });

    it('uses the session fallback when opening IndexedDB fails', async () => {
        const spy = vi.spyOn(indexedDB, 'open').mockImplementation(() => { throw new Error('Unavailable'); });
        allowTestConsole('error', /Error storing secure item/);
        await setSecureItem('token', 'fallback');
        expect(await getSecureItem('token')).toBe('fallback');
        allowTestConsole('error', /Error removing secure item/);
        await removeSecureItem('token');
        allowTestConsole('error', /Error clearing secure storage/);
        await clearSecureStorage();
        spy.mockRestore();
        expect(sessionStorage.getItem('secure_token')).toBeNull();
    });
});
