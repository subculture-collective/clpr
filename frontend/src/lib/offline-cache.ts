/**
 * Offline Cache Layer
 * 
 * Provides normalized caching with IndexedDB for offline-first data access
 * Features:
 * - Normalized entity storage (clips, comments, users by ID)
 * - Feed list caching with entity references
 * - Cache expiration and versioning
 * - Optimistic updates
 */

import { openDB, type DBSchema, type IDBPDatabase } from 'idb';
import type { Clip } from '@/types/clip';
import type { Comment } from '@/types/comment';

// ============================================================================
// Types and Interfaces
// ============================================================================

export interface CacheMetadata {
  timestamp: number;
  expiresAt: number;
  version: number;
}

export interface CachedEntity<T> {
  id: string;
  data: T;
  metadata: CacheMetadata;
}

export interface FeedCache {
  key: string; // Unique key for this feed configuration (e.g., "clips:sort=hot:timeframe=day")
  entityIds: string[]; // Array of entity IDs in order
  page: number;
  hasMore: boolean;
  total: number;
  metadata: CacheMetadata;
}

export interface OfflineCacheConfig {
  dbName: string;
  version: number;
  defaultTTL: number; // Time to live in milliseconds
}

// IndexedDB Schema
interface ClipperCacheDB extends DBSchema {
  clips: {
    key: string;
    value: CachedEntity<Clip>;
    indexes: { 'by-timestamp': number };
  };
  comments: {
    key: string;
    value: CachedEntity<Comment>;
    indexes: { 'by-timestamp': number; 'by-clip': string };
  };
  feeds: {
    key: string;
    value: FeedCache;
    indexes: { 'by-timestamp': number };
  };
  metadata: {
    key: string;
    value: {
      key: string;
      value: unknown;
      timestamp: number;
    };
  };
}

// ============================================================================
// Configuration
// ============================================================================

const DEFAULT_CONFIG: OfflineCacheConfig = {
  dbName: 'clpr-offline-cache',
  version: 1,
  defaultTTL: 1000 * 60 * 60 * 24, // 24 hours
};

// ============================================================================
// Offline Cache Class
// ============================================================================

export class OfflineCache {
  private db: IDBPDatabase<ClipperCacheDB> | null = null;
  private config: OfflineCacheConfig;
  private initPromise: Promise<void> | null = null;

  constructor(config: Partial<OfflineCacheConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  // ============================================================================
  // Initialization
  // ============================================================================

  public async init(): Promise<void> {
    if (this.db) return;
    
    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = this.initializeDB();
    return this.initPromise;
  }

  private async initializeDB(): Promise<void> {
    try {
      this.db = await openDB<ClipperCacheDB>(this.config.dbName, this.config.version, {
        upgrade(db, oldVersion, newVersion) {
          console.log(`[OfflineCache] Upgrading database from version ${oldVersion} to ${newVersion}`);

          // Create clips store
          if (!db.objectStoreNames.contains('clips')) {
            const clipStore = db.createObjectStore('clips', { keyPath: 'id' });
            clipStore.createIndex('by-timestamp', 'metadata.timestamp');
          }

          // Create comments store
          if (!db.objectStoreNames.contains('comments')) {
            const commentStore = db.createObjectStore('comments', { keyPath: 'id' });
            commentStore.createIndex('by-timestamp', 'metadata.timestamp');
            commentStore.createIndex('by-clip', 'data.clip_id');
          }

          // Create feeds store
          if (!db.objectStoreNames.contains('feeds')) {
            const feedStore = db.createObjectStore('feeds', { keyPath: 'key' });
            feedStore.createIndex('by-timestamp', 'metadata.timestamp');
          }

          // Create metadata store
          if (!db.objectStoreNames.contains('metadata')) {
            db.createObjectStore('metadata', { keyPath: 'key' });
          }
        },
        blocked() {
          console.warn('[OfflineCache] Database upgrade blocked. Please close other tabs.');
        },
        blocking() {
          console.warn('[OfflineCache] This connection is blocking a database upgrade.');
        },
      });

      console.log('[OfflineCache] Database initialized successfully');
    } catch (error) {
      console.error('[OfflineCache] Failed to initialize database:', error);
      throw error;
    }
  }

  private async ensureDB(): Promise<IDBPDatabase<ClipperCacheDB>> {
    if (!this.db) {
      await this.init();
    }
    if (!this.db) {
      throw new Error('Database not initialized');
    }
    return this.db;
  }

  // ============================================================================
  // Clip Operations
  // ============================================================================

  public async getClip(id: string): Promise<Clip | null> {
    const db = await this.ensureDB();
    const cached = await db.get('clips', id);
    
    if (!cached) return null;
    
    // Check if expired
    if (Date.now() > cached.metadata.expiresAt) {
      await this.deleteClip(id);
      return null;
    }
    
    return cached.data;
  }

  public async setClip(clip: Clip, ttl?: number): Promise<void> {
    const db = await this.ensureDB();
    const now = Date.now();
    const expiresAt = now + (ttl || this.config.defaultTTL);
    
    const cached: CachedEntity<Clip> = {
      id: clip.id,
      data: clip,
      metadata: {
        timestamp: now,
        expiresAt,
        version: this.config.version,
      },
    };
    
    await db.put('clips', cached);
  }

  public async setClips(clips: Clip[], ttl?: number): Promise<void> {
    const db = await this.ensureDB();
    const tx = db.transaction('clips', 'readwrite');
    const now = Date.now();
    const expiresAt = now + (ttl || this.config.defaultTTL);
    
    await Promise.all([
      ...clips.map(clip =>
        tx.store.put({
          id: clip.id,
          data: clip,
          metadata: {
            timestamp: now,
            expiresAt,
            version: this.config.version,
          },
        })
      ),
      tx.done,
    ]);
  }

  public async deleteClip(id: string): Promise<void> {
    const db = await this.ensureDB();
    await db.delete('clips', id);
  }

  // ============================================================================
  // Comment Operations
  // ============================================================================

  public async getComment(id: string): Promise<Comment | null> {
    const db = await this.ensureDB();
    const cached = await db.get('comments', id);
    
    if (!cached) return null;
    
    // Check if expired
    if (Date.now() > cached.metadata.expiresAt) {
      await this.deleteComment(id);
      return null;
    }
    
    return cached.data;
  }

  public async getCommentsByClipId(clipId: string): Promise<Comment[]> {
    const db = await this.ensureDB();
    const cached = await db.getAllFromIndex('comments', 'by-clip', clipId);
    const now = Date.now();
    
    // Filter out expired comments
    const validComments = cached.filter(c => now <= c.metadata.expiresAt);
    
    return validComments.map(c => c.data);
  }

  public async setComment(comment: Comment, ttl?: number): Promise<void> {
    const db = await this.ensureDB();
    const now = Date.now();
    const expiresAt = now + (ttl || this.config.defaultTTL);
    
    const cached: CachedEntity<Comment> = {
      id: comment.id,
      data: comment,
      metadata: {
        timestamp: now,
        expiresAt,
        version: this.config.version,
      },
    };
    
    await db.put('comments', cached);
  }

  public async setComments(comments: Comment[], ttl?: number): Promise<void> {
    const db = await this.ensureDB();
    const tx = db.transaction('comments', 'readwrite');
    const now = Date.now();
    const expiresAt = now + (ttl || this.config.defaultTTL);
    
    await Promise.all([
      ...comments.map(comment =>
        tx.store.put({
          id: comment.id,
          data: comment,
          metadata: {
            timestamp: now,
            expiresAt,
            version: this.config.version,
          },
        })
      ),
      tx.done,
    ]);
  }

  public async deleteComment(id: string): Promise<void> {
    const db = await this.ensureDB();
    await db.delete('comments', id);
  }

  // ============================================================================
  // Feed Operations
  // ============================================================================

  public async getFeed(key: string): Promise<FeedCache | null> {
    const db = await this.ensureDB();
    const cached = await db.get('feeds', key);
    
    if (!cached) return null;
    
    // Check if expired
    if (Date.now() > cached.metadata.expiresAt) {
      await this.deleteFeed(key);
      return null;
    }
    
    return cached;
  }

  public async setFeed(feed: FeedCache, ttl?: number): Promise<void> {
    const db = await this.ensureDB();
    const now = Date.now();
    const expiresAt = now + (ttl || this.config.defaultTTL);
    
    const cached: FeedCache = {
      ...feed,
      metadata: {
        timestamp: now,
        expiresAt,
        version: this.config.version,
      },
    };
    
    await db.put('feeds', cached);
  }

  public async deleteFeed(key: string): Promise<void> {
    const db = await this.ensureDB();
    await db.delete('feeds', key);
  }

  // ============================================================================
  // Metadata Operations
  // ============================================================================

  public async getMetadata<T>(key: string): Promise<T | null> {
    const db = await this.ensureDB();
    const item = await db.get('metadata', key);
    return item ? (item.value as T) : null;
  }

  public async setMetadata<T>(key: string, value: T): Promise<void> {
    const db = await this.ensureDB();
    await db.put('metadata', {
      key,
      value,
      timestamp: Date.now(),
    });
  }

  // ============================================================================
  // Utility Operations
  // ============================================================================

  public async clearExpired(): Promise<void> {
    const db = await this.ensureDB();
    const now = Date.now();
    
    // Clear expired clips
    const clips = await db.getAll('clips');
    const expiredClips = clips.filter(c => now > c.metadata.expiresAt);
    await Promise.all(expiredClips.map(c => db.delete('clips', c.id)));
    
    // Clear expired comments
    const comments = await db.getAll('comments');
    const expiredComments = comments.filter(c => now > c.metadata.expiresAt);
    await Promise.all(expiredComments.map(c => db.delete('comments', c.id)));
    
    // Clear expired feeds
    const feeds = await db.getAll('feeds');
    const expiredFeeds = feeds.filter(f => now > f.metadata.expiresAt);
    await Promise.all(expiredFeeds.map(f => db.delete('feeds', f.key)));
    
    console.log(`[OfflineCache] Cleared ${expiredClips.length} clips, ${expiredComments.length} comments, ${expiredFeeds.length} feeds`);
  }

  public async clear(): Promise<void> {
    const db = await this.ensureDB();
    await Promise.all([
      db.clear('clips'),
      db.clear('comments'),
      db.clear('feeds'),
      db.clear('metadata'),
    ]);
    console.log('[OfflineCache] All caches cleared');
  }

  public async getStats(): Promise<{
    clips: number;
    comments: number;
    feeds: number;
  }> {
    const db = await this.ensureDB();
    const [clips, comments, feeds] = await Promise.all([
      db.count('clips'),
      db.count('comments'),
      db.count('feeds'),
    ]);
    
    return { clips, comments, feeds };
  }

  public async close(): Promise<void> {
    if (this.db) {
      this.db.close();
      this.db = null;
      this.initPromise = null;
    }
  }
}

// ============================================================================
// Singleton Instance
// ============================================================================

let offlineCacheInstance: OfflineCache | null = null;

export function getOfflineCache(): OfflineCache {
  if (!offlineCacheInstance) {
    offlineCacheInstance = new OfflineCache();
  }
  return offlineCacheInstance;
}

export function resetOfflineCache(): void {
  if (offlineCacheInstance) {
    offlineCacheInstance.close();
    offlineCacheInstance = null;
  }
}

// Export default instance
export const offlineCache = getOfflineCache();
