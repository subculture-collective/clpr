/**
 * Session-scoped browser storage for authentication data.
 *
 * IndexedDB values are encrypted to reduce disclosure from an offline database
 * copy, but the exportable key lives in the same browser origin/session. This is
 * not an XSS security boundary: injected same-origin JavaScript can access both
 * the key and plaintext. Server-side expiry, CSP, HttpOnly cookies where used,
 * and logout cleanup remain the actual controls.
 */

const DB_NAME = 'clpr-secure-storage';
const STORE_NAME = 'encrypted-data';
const DB_VERSION = 1;

// Encryption key storage key
const ENCRYPTION_KEY_NAME = 'clpr-encryption-key';

/**
 * Initialize IndexedDB
 */
async function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    };
  });
}

/**
 * Generate or retrieve encryption key
 */
async function getEncryptionKey(): Promise<CryptoKey> {
  // Check if we have a key in sessionStorage (ephemeral)
  const storedKey = sessionStorage.getItem(ENCRYPTION_KEY_NAME);

  if (storedKey) {
    try {
      const keyData = JSON.parse(storedKey);
      return await crypto.subtle.importKey(
        'jwk',
        keyData,
        { name: 'AES-GCM', length: 256 },
        true,
        ['encrypt', 'decrypt']
      );
    } catch {
      // Corrupted key data - remove and generate fresh
      sessionStorage.removeItem(ENCRYPTION_KEY_NAME);
    }
  }

  // Generate new key
  const key = await crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    true,
    ['encrypt', 'decrypt']
  );

  // Export and store key
  const exportedKey = await crypto.subtle.exportKey('jwk', key);
  sessionStorage.setItem(ENCRYPTION_KEY_NAME, JSON.stringify(exportedKey));

  return key;
}

/**
 * Encrypt data using Web Crypto API
 */
async function encryptData(data: string): Promise<{ iv: Uint8Array; ciphertext: ArrayBuffer }> {
  const key = await getEncryptionKey();
  const encoder = new TextEncoder();
  const iv = crypto.getRandomValues(new Uint8Array(12)); // GCM recommends 12 bytes

  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    encoder.encode(data)
  );

  return { iv, ciphertext };
}

/**
 * Decrypt data using Web Crypto API
 */
async function decryptData(iv: Uint8Array, ciphertext: ArrayBuffer): Promise<string> {
  const key = await getEncryptionKey();
  const decoder = new TextDecoder();

  const plaintext = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: new Uint8Array(iv) },
    key,
    ciphertext
  );

  return decoder.decode(plaintext);
}

/**
 * Check if secure storage is available
 */
export function isSecureStorageAvailable(): boolean {
  return typeof indexedDB !== 'undefined' && typeof crypto.subtle !== 'undefined';
}

/**
 * Helper to clear all secure_* prefixed keys from a storage object
 */
function clearSecurePrefixedKeys(storage: Storage): void {
  const keysToRemove: string[] = [];
  for (let i = 0; i < storage.length; i++) {
    const key = storage.key(i);
    if (key?.startsWith('secure_')) {
      keysToRemove.push(key);
    }
  }
  keysToRemove.forEach(key => storage.removeItem(key));
}

// Resolve writes after commit, not merely after the individual request succeeds.
// An aborted transaction must follow the same fallback path as an open failure.
async function storeRequest<T>(
  mode: IDBTransactionMode,
  operation: (store: IDBObjectStore) => IDBRequest<T>,
): Promise<T> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    try {
      const transaction = db.transaction(STORE_NAME, mode);
      const request = operation(transaction.objectStore(STORE_NAME));
      transaction.oncomplete = () => {
        db.close();
        resolve(request.result);
      };
      transaction.onabort = () => {
        db.close();
        reject(transaction.error ?? new Error('Secure storage transaction aborted'));
      };
    } catch (error) {
      db.close();
      reject(error);
    }
  });
}

export async function setSecureItem(key: string, value: string): Promise<void> {
  if (!isSecureStorageAvailable()) {
    sessionStorage.setItem('secure_' + key, value);
    return;
  }
  try {
    const { iv, ciphertext } = await encryptData(value);
    await storeRequest('readwrite', store => store.put({
      iv: Array.from(iv),
      ciphertext: Array.from(new Uint8Array(ciphertext)),
    }, key));
    sessionStorage.removeItem('secure_' + key);
  } catch (error) {
    console.error('Error storing secure item:', error);
    sessionStorage.setItem('secure_' + key, value);
  }
}

export async function getSecureItem(key: string): Promise<string | null> {
  const fallback = () => sessionStorage.getItem('secure_' + key);
  if (fallback() !== null || !isSecureStorageAvailable()) return fallback();
  try {
    const data = await storeRequest('readonly', store => store.get(key));
    if (!data) return fallback();
    return await decryptData(new Uint8Array(data.iv), new Uint8Array(data.ciphertext).buffer);
  } catch (error) {
    console.error('Error retrieving secure item:', error);
    return fallback();
  }
}

export async function removeSecureItem(key: string): Promise<void> {
  localStorage.removeItem('secure_' + key);
  sessionStorage.removeItem('secure_' + key);
  if (!isSecureStorageAvailable()) return;
  try {
    await storeRequest('readwrite', store => store.delete(key));
  } catch (error) {
    console.error('Error removing secure item:', error);
  }
}

export async function clearSecureStorage(): Promise<void> {
  clearSecurePrefixedKeys(sessionStorage);
  clearSecurePrefixedKeys(localStorage);
  if (!isSecureStorageAvailable()) return;
  try {
    await storeRequest('readwrite', store => store.clear());
  } catch (error) {
    console.error('Error clearing secure storage:', error);
  }
}
