import { apiClient } from './api';
import type {
    WebhookSubscription,
    WebhookDelivery,
    CreateWebhookSubscriptionRequest,
    UpdateWebhookSubscriptionRequest,
    WebhookEventCatalog,
} from '../types/webhook';

// Get supported webhook events
export async function getSupportedWebhookEvents(): Promise<WebhookEventCatalog> {
    const response = await apiClient.get('/webhooks/events');
    return {
        events: response.data.data || [],
        discordEvents: response.data.meta?.discord_events || [],
    };
}

// Create a new webhook subscription
export async function createWebhookSubscription(
    data: CreateWebhookSubscriptionRequest
): Promise<{ subscription: WebhookSubscription; secret: string }> {
    const response = await apiClient.post('/webhooks', data);
    return {
        subscription: response.data.data,
        secret: response.data.secret,
    };
}

// List all webhook subscriptions
export async function listWebhookSubscriptions(): Promise<
    WebhookSubscription[]
> {
    const response = await apiClient.get('/webhooks');
    return response.data.data;
}

// Get a specific webhook subscription
export async function getWebhookSubscription(
    id: string
): Promise<WebhookSubscription> {
    const response = await apiClient.get(`/webhooks/${id}`);
    return response.data.data;
}

// Update a webhook subscription
// NOTE: Some backends may respond to PATCH /webhooks/:id with a 204 No Content
// or without including the updated subscription in response.data.data.
// To maintain this helper's contract of always returning a WebhookSubscription,
// we perform a follow-up GET only when the PATCH response does not contain data.
// This extra network request has a performance cost but is intentional until
// the backend can be updated to always return the updated resource.
// If the backend behavior changes, this fallback can be removed.
export async function updateWebhookSubscription(
    id: string,
    data: UpdateWebhookSubscriptionRequest
): Promise<WebhookSubscription> {
    const response = await apiClient.patch(`/webhooks/${id}`, data);
    // Prefer the subscription data from the PATCH response when available
    if (response.data.data) {
        return response.data.data;
    }
    // Intentional fallback: fetch the updated subscription when the backend
    // does not return it in the PATCH response
    return getWebhookSubscription(id);
}

// Delete a webhook subscription
export async function deleteWebhookSubscription(id: string): Promise<void> {
    await apiClient.delete(`/webhooks/${id}`);
}

// Regenerate webhook secret
export async function regenerateWebhookSecret(
    id: string
): Promise<string> {
    const response = await apiClient.post(`/webhooks/${id}/regenerate-secret`);
    return response.data.secret;
}

// Get delivery history for a webhook subscription
export async function getWebhookDeliveries(
    id: string,
    page: number = 1,
    limit: number = 20
): Promise<{
    deliveries: WebhookDelivery[];
    meta: {
        page: number;
        limit: number;
        total: number;
        total_pages: number;
    };
}> {
    const response = await apiClient.get(`/webhooks/${id}/deliveries`, {
        params: { page, limit },
    });
    return {
        deliveries: response.data.data,
        meta: response.data.meta,
    };
}

// Admin: Get webhook dead-letter queue items
export async function getWebhookDLQItems(
    page: number = 1,
    limit: number = 20
): Promise<{
    items: import('../types/webhook').OutboundWebhookDLQItem[];
    pagination: {
        page: number;
        limit: number;
        total: number;
        total_pages: number;
    };
}> {
    const response = await apiClient.get('/admin/webhooks/dlq', {
        params: { page, limit },
    });
    return response.data;
}

// Admin: Replay a webhook DLQ item
export async function replayWebhookDLQItem(id: string): Promise<void> {
    await apiClient.post(`/admin/webhooks/dlq/${id}/replay`);
}

// Admin: Delete a webhook DLQ item
export async function deleteWebhookDLQItem(id: string): Promise<void> {
    await apiClient.delete(`/admin/webhooks/dlq/${id}`);
}
