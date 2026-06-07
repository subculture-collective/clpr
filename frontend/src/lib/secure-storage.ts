/**
 * Secure Storage for sensitive data in PWA
 * Uses IndexedDB with Web Crypto API for encryption on mobile/PWA
 * Falls back to sessionStorage for environments without crypto support
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
      return crypto.subtle.importKey(
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
    { name: 'AES-GCM', iv },
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

/**
 * Store encrypted data
 */
export async function setSecureItem(key: string, value: string): Promise<void> {
  if (!isSecureStorageAvailable()) {
    // Fallback to sessionStorage only (not localStorage) to limit exposure
    sessionStorage.setItem(`secure_${key}`, value);
    return;
  }

  try {
    const { iv, ciphertext } = await encryptData(value);
    const db = await openDB();

    return new Promise((resolve, reject) => {
      const transaction = db.transaction(STORE_NAME, 'readwrite');
      const store = transaction.objectStore(STORE_NAME);

      // Store IV and ciphertext
      const data = {
        iv: Array.from(iv), // Convert to array for storage
        ciphertext: Array.from(new Uint8Array(ciphertext)),
      };

      const request = store.put(data, key);

      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);

      transaction.oncomplete = () => db.close();
    });
  } catch (error) {
    console.error('Error storing secure item:', error);
    // Fallback to sessionStorage only
    sessionStorage.setItem(`secure_${key}`, value);
  }
}

/**
 * Retrieve and decrypt data
 */
export async function getSecureItem(key: string): Promise<string | null> {
  if (!isSecureStorageAvailable()) {
    return sessionStorage.getItem(`secure_${key}`);
  }

  try {
    const db = await openDB();

    return new Promise((resolve, reject) => {
      const transaction = db.transaction(STORE_NAME, 'readonly');
      const store = transaction.objectStore(STORE_NAME);
      const request = store.get(key);

      request.onsuccess = async () => {
        const data = request.result;
        if (!data) {
          // Check sessionStorage fallback
          resolve(sessionStorage.getItem(`secure_${key}`));
          return;
        }

        try {
          const iv = new Uint8Array(data.iv);
          const ciphertext = new Uint8Array(data.ciphertext).buffer;
          const plaintext = await decryptData(iv, ciphertext);
          resolve(plaintext);
        } catch (error) {
          console.error('Error decrypting data:', error);
          resolve(sessionStorage.getItem(`secure_${key}`));
        }
      };

      request.onerror = () => reject(request.error);

      transaction.oncomplete = () => db.close();
    });
  } catch (error) {
    console.error('Error retrieving secure item:', error);
    return sessionStorage.getItem(`secure_${key}`);
  }
}

/**
 * Remove encrypted data
 */
export async function removeSecureItem(key: string): Promise<void> {
  // Always clean up both storages
  localStorage.removeItem(`secure_${key}`);
  sessionStorage.removeItem(`secure_${key}`);

  if (!isSecureStorageAvailable()) {
    return;
  }

  try {
    const db = await openDB();

    return new Promise((resolve, reject) => {
      const transaction = db.transaction(STORE_NAME, 'readwrite');
      const store = transaction.objectStore(STORE_NAME);
      const request = store.delete(key);

      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);

      transaction.oncomplete = () => db.close();
    });
  } catch (error) {
    console.error('Error removing secure item:', error);
  }
}

/**
 * Clear all secure storage
 */
export async function clearSecureStorage(): Promise<void> {
  // Always clear both storages regardless of IndexedDB availability
  clearSecurePrefixedKeys(sessionStorage);
  clearSecurePrefixedKeys(localStorage);

  if (!isSecureStorageAvailable()) {
    return;
  }

  try {
    const db = await openDB();

    return new Promise((resolve, reject) => {
      const transaction = db.transaction(STORE_NAME, 'readwrite');
      const store = transaction.objectStore(STORE_NAME);
      const request = store.clear();

      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);

      transaction.oncomplete = () => db.close();
    });
  } catch (error) {
    console.error('Error clearing secure storage:', error);
  }
}
