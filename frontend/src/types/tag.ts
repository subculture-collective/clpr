/** Which process produced a tag; see backend/internal/tagtaxonomy. */
export type TagLane =
  | "category"
  | "detected"
  | "streamer"
  | "community"
  | "duration"
  | "language"
  | "root";

/** Evidence a detected tag's definition requires (not a per-clip verdict). */
export type TagEvidence = "visible" | "contextual" | "strong";

export interface Tag {
  id: string;
  name: string;
  slug: string;
  parent_slug?: string | null;
  description?: string;
  color?: string;
  usage_count: number;
  created_at: string;
  suppressed_at?: string;
  suppressed_by?: string;
  suppression_reason?: string;
  lane?: TagLane;
  display_name?: string;
  evidence?: TagEvidence;
}

export interface TagTreeNode extends Tag {
  children?: TagTreeNode[];
}

export interface TagTreeResponse {
  tags: TagTreeNode[];
}

export interface TagListResponse {
  tags: Tag[];
  total: number;
  page: number;
  limit: number;
  has_more: boolean;
}

export interface TagSearchResponse {
  tags: Tag[];
}

export interface TagDetailResponse {
  tag: Tag;
  clip_count: number;
}

export interface ClipTagsResponse {
  tags: Tag[];
}

export interface AddTagsRequest {
  tag_slugs: string[];
}

export interface TagPromotionItem {
  id: string;
  tag_slug: string;
  unique_users: number;
  usage_count: number;
  status: "pending" | "approved" | "rejected";
  created_at: string;
}

export interface TagPromotionQueueResponse {
  items: TagPromotionItem[];
  total: number;
}
