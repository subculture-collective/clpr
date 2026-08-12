// Category represents a high-level content category
export interface Category {
    id: string;
    name: string;
    slug: string;
    description?: string;
    icon?: string;
    position: number;
    category_type?: string;
    is_featured?: boolean;
    is_custom?: boolean;
    is_public?: boolean;
    created_by_user_id?: string | null;
    created_at: string;
    updated_at: string;
}

// CategoryListResponse represents the response for listing categories
export interface CategoryListResponse {
    categories: Category[];
}

// CategoryDetailResponse represents the response for a single category
export interface CategoryDetailResponse {
    category: Category;
}
