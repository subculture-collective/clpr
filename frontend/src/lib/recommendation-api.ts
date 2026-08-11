import { apiClient } from './api';

export interface RecommendationPreferences {
    user_id: string;
    followed_creators: string[];
    preferred_topics: string[];
    preferred_tags: string[];
    twitch_categories: string[];
    onboarding_completed: boolean;
}

export interface CreatorFirstOnboardingRequest {
    followed_creators: string[];
    preferred_topics: string[];
    preferred_tags: string[];
}

export async function fetchRecommendationPreferences() {
    const response = await apiClient.get<RecommendationPreferences>(
        '/recommendations/preferences',
    );
    return response.data;
}

export async function completeCreatorFirstOnboarding(
    request: CreatorFirstOnboardingRequest,
) {
    const response = await apiClient.post<RecommendationPreferences>(
        '/recommendations/onboarding',
        request,
    );
    return response.data;
}
