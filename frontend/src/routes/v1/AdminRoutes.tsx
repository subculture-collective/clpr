import type { ComponentType } from 'react';
import { Route } from 'react-router-dom';
import { AdminRoute } from '@/components/guards/AdminRoute';

type Page = ComponentType;

export interface AdminRoutePages {
    dashboard: Page;
    clips: Page;
    comments: Page;
    users: Page;
    reports: Page;
    webhookDlq: Page;
    sync: Page;
    analytics: Page;
    revenue: Page;
    campaigns: Page;
    submissions: Page;
    moderation: Page;
    moderationAnalytics: Page;
    moderators: Page;
    bans: Page;
    auditLogs: Page;
    verification: Page;
    discoveryLists: Page;
    discoveryListForm: Page;
    playlistScripts: Page;
    tags: Page;
    topics: Page;
    apiDocs: Page;
    forumModeration: Page;
    forumModerationLog: Page;
    moderationUsers: Page;
    tagPromotion: Page;
}

function protectedPage(PageComponent: Page) {
    return (
        <AdminRoute>
            <PageComponent />
        </AdminRoute>
    );
}

/** Versioned administration route boundary. Every route is fail-closed by AdminRoute. */
export function adminRoutes(pages: AdminRoutePages) {
    return (
        <>
            <Route path='/admin/dashboard' element={protectedPage(pages.dashboard)} />
            <Route path='/admin/clips' element={protectedPage(pages.clips)} />
            <Route path='/admin/comments' element={protectedPage(pages.comments)} />
            <Route path='/admin/users' element={protectedPage(pages.users)} />
            <Route path='/admin/reports' element={protectedPage(pages.reports)} />
            <Route path='/admin/webhooks/dlq' element={protectedPage(pages.webhookDlq)} />
            <Route path='/admin/sync' element={protectedPage(pages.sync)} />
            <Route path='/admin/analytics' element={protectedPage(pages.analytics)} />
            <Route path='/admin/revenue' element={protectedPage(pages.revenue)} />
            <Route path='/admin/campaigns' element={protectedPage(pages.campaigns)} />
            <Route path='/admin/submissions' element={protectedPage(pages.submissions)} />
            <Route path='/admin/moderation' element={protectedPage(pages.moderation)} />
            <Route
                path='/admin/moderation/analytics'
                element={protectedPage(pages.moderationAnalytics)}
            />
            <Route path='/admin/moderators' element={protectedPage(pages.moderators)} />
            <Route path='/admin/bans' element={protectedPage(pages.bans)} />
            <Route path='/admin/audit-logs' element={protectedPage(pages.auditLogs)} />
            <Route path='/admin/verification' element={protectedPage(pages.verification)} />
            <Route
                path='/admin/discovery-lists'
                element={protectedPage(pages.discoveryLists)}
            />
            <Route
                path='/admin/discovery-lists/:id/edit'
                element={protectedPage(pages.discoveryListForm)}
            />
            <Route
                path='/admin/playlist-scripts'
                element={protectedPage(pages.playlistScripts)}
            />
            <Route path='/admin/tags' element={protectedPage(pages.tags)} />
            <Route path='/admin/topics' element={protectedPage(pages.topics)} />
            <Route
                path='/admin/tag-promotion'
                element={protectedPage(pages.tagPromotion)}
            />
            <Route path='/admin/api-docs' element={protectedPage(pages.apiDocs)} />
            <Route
                path='/admin/forum/moderation'
                element={protectedPage(pages.forumModeration)}
            />
            <Route
                path='/admin/forum/moderation-log'
                element={protectedPage(pages.forumModerationLog)}
            />
            <Route path='/moderation/users' element={protectedPage(pages.moderationUsers)} />
        </>
    );
}
