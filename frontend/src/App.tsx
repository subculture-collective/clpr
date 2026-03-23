import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { HelmetProvider } from '@dr.pogodin/react-helmet';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider } from './context/ToastContext';
import { ConsentProvider } from './context/ConsentContext';
import { AppLayout } from './components/layout';
import { ProtectedRoute, AdminRoute, GuestRoute } from './components/guards';
import { Spinner } from './components';
import { ConsentBanner } from './components/consent';

// Lazy load page components for code splitting
const HomePage = lazy(() =>
    import('./pages/HomePage').then(m => ({ default: m.HomePage })),
);
const DiscoveryListsPage = lazy(() =>
    import('./pages/DiscoveryListsPage').then(m => ({
        default: m.DiscoveryListsPage,
    })),
);
const DiscoveryListDetailPage = lazy(() =>
    import('./pages/DiscoveryListDetailPage').then(m => ({
        default: m.DiscoveryListDetailPage,
    })),
);
const ScrapedClipsPage = lazy(() =>
    import('./pages/ScrapedClipsPage').then(m => ({
        default: m.ScrapedClipsPage,
    })),
);
const LiveFeedPage = lazy(() =>
    import('./pages/LiveFeedPage').then(m => ({ default: m.LiveFeedPage })),
);
const ClipDetailPage = lazy(() =>
    import('./pages/ClipDetailPage').then(m => ({ default: m.ClipDetailPage })),
);
const GamePage = lazy(() =>
    import('./pages/GamePage').then(m => ({ default: m.GamePage })),
);
const CategoryPage = lazy(() =>
    import('./pages/CategoryPage').then(m => ({ default: m.CategoryPage })),
);
const BroadcasterPage = lazy(() =>
    import('./pages/BroadcasterPage').then(m => ({
        default: m.BroadcasterPage,
    })),
);
const CreatorPage = lazy(() =>
    import('./pages/CreatorPage').then(m => ({
        default: m.CreatorPage,
    })),
);
const UserProfilePage = lazy(() =>
    import('./pages/UserProfilePage').then(m => ({
        default: m.UserProfilePage,
    })),
);
const TagPage = lazy(() =>
    import('./pages/TagPage').then(m => ({ default: m.TagPage })),
);
const SearchPage = lazy(() =>
    import('./pages/SearchPage').then(m => ({ default: m.SearchPage })),
);
const AboutPage = lazy(() =>
    import('./pages/AboutPage').then(m => ({ default: m.AboutPage })),
);
const PrivacyPage = lazy(() =>
    import('./pages/PrivacyPage').then(m => ({ default: m.PrivacyPage })),
);
const TermsPage = lazy(() =>
    import('./pages/TermsPage').then(m => ({ default: m.TermsPage })),
);
const DMCAPage = lazy(() =>
    import('./pages/DMCAPage').then(m => ({ default: m.DMCAPage })),
);
const CommunityRulesPage = lazy(() =>
    import('./pages/CommunityRulesPage').then(m => ({
        default: m.CommunityRulesPage,
    })),
);
const ContactPage = lazy(() =>
    import('./pages/ContactPage').then(m => ({ default: m.ContactPage })),
);
const DocsPage = lazy(() =>
    import('./pages/DocsPage').then(m => ({ default: m.DocsPage })),
);
const ExtensionPage = lazy(() =>
    import('./pages/ExtensionPage').then(m => ({ default: m.ExtensionPage })),
);
const NotFoundPage = lazy(() =>
    import('./pages/NotFoundPage').then(m => ({ default: m.NotFoundPage })),
);
const LoginPage = lazy(() =>
    import('./pages/LoginPage').then(m => ({ default: m.LoginPage })),
);
const AuthCallbackPage = lazy(() =>
    import('./pages/AuthCallbackPage').then(m => ({
        default: m.AuthCallbackPage,
    })),
);
const FavoritesPage = lazy(() =>
    import('./pages/FavoritesPage').then(m => ({ default: m.FavoritesPage })),
);
const ProfilePage = lazy(() =>
    import('./pages/ProfilePage').then(m => ({ default: m.ProfilePage })),
);
const SubmitClipPage = lazy(() =>
    import('./pages/SubmitClipPage').then(m => ({ default: m.SubmitClipPage })),
);
const UserSubmissionsPage = lazy(() =>
    import('./pages/UserSubmissionsPage').then(m => ({
        default: m.UserSubmissionsPage,
    })),
);
const SettingsPage = lazy(() =>
    import('./pages/SettingsPage').then(m => ({ default: m.SettingsPage })),
);
const CookieSettingsPage = lazy(() =>
    import('./pages/CookieSettingsPage').then(m => ({
        default: m.CookieSettingsPage,
    })),
);
const CreatorDashboardPage = lazy(() =>
    import('./pages/CreatorDashboardPage').then(m => ({
        default: m.CreatorDashboardPage,
    })),
);
const AdminDashboard = lazy(() =>
    import('./pages/admin/AdminDashboard').then(m => ({
        default: m.AdminDashboard,
    })),
);
const AdminClipsPage = lazy(() =>
    import('./pages/admin/AdminClipsPage').then(m => ({
        default: m.AdminClipsPage,
    })),
);
const AdminCommentsPage = lazy(() =>
    import('./pages/admin/AdminCommentsPage').then(m => ({
        default: m.AdminCommentsPage,
    })),
);
const AdminUsersPage = lazy(() =>
    import('./pages/admin/AdminUsersPage').then(m => ({
        default: m.AdminUsersPage,
    })),
);
const AdminReportsPage = lazy(() =>
    import('./pages/admin/AdminReportsPage').then(m => ({
        default: m.AdminReportsPage,
    })),
);
const AdminWebhookDLQPage = lazy(() =>
    import('./pages/admin/AdminWebhookDLQPage').then(m => ({
        default: m.AdminWebhookDLQPage,
    })),
);
const AdminSyncPage = lazy(() =>
    import('./pages/admin/AdminSyncPage').then(m => ({
        default: m.AdminSyncPage,
    })),
);
const ModerationQueuePage = lazy(() =>
    import('./pages/admin/ModerationQueuePage').then(m => ({
        default: m.ModerationQueuePage,
    })),
);
const AdminModerationQueuePage = lazy(() =>
    import('./pages/admin/AdminModerationQueuePage').then(m => ({
        default: m.AdminModerationQueuePage,
    })),
);
const AdminVerificationQueuePage = lazy(() =>
    import('./pages/admin/AdminVerificationQueuePage').then(m => ({
        default: m.AdminVerificationQueuePage,
    })),
);
const AdminModerationAnalyticsPage = lazy(
    () => import('./pages/admin/AdminModerationAnalyticsPage'),
);
const AdminModeratorsPage = lazy(() =>
    import('./pages/admin/AdminModeratorsPage').then(m => ({
        default: m.AdminModeratorsPage,
    })),
);
const AdminBansPage = lazy(() =>
    import('./pages/admin/AdminBansPage').then(m => ({
        default: m.AdminBansPage,
    })),
);
const AdminAuditLogsPage = lazy(() =>
    import('./pages/admin/AdminAuditLogsPage').then(m => ({
        default: m.AdminAuditLogsPage,
    })),
);
const LeaderboardPage = lazy(() => import('./pages/LeaderboardPage'));
const NotificationsPage = lazy(() =>
    import('./pages/NotificationsPage').then(m => ({
        default: m.NotificationsPage,
    })),
);
const NotificationPreferencesPage = lazy(() =>
    import('./pages/NotificationPreferencesPage').then(m => ({
        default: m.NotificationPreferencesPage,
    })),
);
const CreatorAnalyticsPage = lazy(() => import('./pages/CreatorAnalyticsPage'));
const PersonalStatsPage = lazy(() => import('./pages/PersonalStatsPage'));
const AdminAnalyticsPage = lazy(
    () => import('./pages/admin/AdminAnalyticsPage'),
);
const AdminRevenuePage = lazy(() => import('./pages/admin/AdminRevenuePage'));
const AdminCampaignsPage = lazy(
    () => import('./pages/admin/AdminCampaignsPage'),
);
const AdminDiscoveryListsPage = lazy(() =>
    import('./pages/admin/AdminDiscoveryListsPage').then(m => ({
        default: m.AdminDiscoveryListsPage,
    })),
);
const AdminDiscoveryListFormPage = lazy(() =>
    import('./pages/admin/AdminDiscoveryListFormPage').then(m => ({
        default: m.AdminDiscoveryListFormPage,
    })),
);
const AdminPlaylistScriptsPage = lazy(() =>
    import('./pages/admin/AdminPlaylistScriptsPage').then(m => ({
        default: m.AdminPlaylistScriptsPage,
    })),
);
const AdminTagsPage = lazy(() =>
    import('./pages/admin/AdminTagsPage').then(m => ({
        default: m.AdminTagsPage,
    })),
);
const AdminAPIDocsPage = lazy(() =>
    import('./pages/admin/AdminAPIDocsPage').then(m => ({
        default: m.AdminAPIDocsPage,
    })),
);
const PricingPage = lazy(() => import('./pages/PricingPage'));
const SubscriptionSuccessPage = lazy(
    () => import('./pages/SubscriptionSuccessPage'),
);
const SubscriptionCancelPage = lazy(
    () => import('./pages/SubscriptionCancelPage'),
);
const RoleBadgeTestPage = lazy(() =>
    import('./pages/RoleBadgeTestPage').then(m => ({
        default: m.RoleBadgeTestPage,
    })),
);
const VerifiedBadgeTestPage = lazy(() =>
    import('./pages/VerifiedBadgeTestPage').then(m => ({
        default: m.VerifiedBadgeTestPage,
    })),
);
const VerificationApplicationPage = lazy(() =>
    import('./pages/VerificationApplicationPage').then(m => ({
        default: m.VerificationApplicationPage,
    })),
);
const PlaylistsPage = lazy(() =>
    import('./pages/PlaylistsPage').then(m => ({ default: m.PlaylistsPage })),
);
const PlaylistCreatePage = lazy(() =>
    import('./pages/PlaylistCreatePage').then(m => ({
        default: m.PlaylistCreatePage,
    })),
);
const PlaylistDetailPage = lazy(() =>
    import('./pages/PlaylistDetailPage').then(m => ({
        default: m.PlaylistDetailPage,
    })),
);
const PlaylistTheatrePage = lazy(() =>
    import('./pages/PlaylistTheatrePage').then(m => ({
        default: m.PlaylistTheatrePage,
    })),
);
const PublicPlaylistsPage = lazy(() =>
    import('./pages/PublicPlaylistsPage').then(m => ({
        default: m.PublicPlaylistsPage,
    })),
);
const BookmarkedPlaylistsPage = lazy(() =>
    import('./pages/BookmarkedPlaylistsPage').then(m => ({
        default: m.BookmarkedPlaylistsPage,
    })),
);
const SmartPlaylistsPage = lazy(() =>
    import('./pages/SmartPlaylistsPage').then(m => ({
        default: m.SmartPlaylistsPage,
    })),
);
const WatchHistoryPage = lazy(() =>
    import('./pages/WatchHistoryPage').then(m => ({
        default: m.WatchHistoryPage,
    })),
);
const QueuePage = lazy(() =>
    import('./pages/QueuePage').then(m => ({ default: m.QueuePage })),
);
const QueueTheatrePage = lazy(() =>
    import('./pages/QueueTheatrePage').then(m => ({
        default: m.QueueTheatrePage,
    })),
);
const StreamPage = lazy(() =>
    import('./pages/StreamPage').then(m => ({ default: m.StreamPage })),
);
const ForumModerationPage = lazy(() =>
    import('./pages/admin/ForumModerationPage').then(m => ({
        default: m.ForumModerationPage,
    })),
);
const ModerationLogPage = lazy(() =>
    import('./pages/admin/ModerationLogPage').then(m => ({
        default: m.ModerationLogPage,
    })),
);
const ModerationUsersPage = lazy(() =>
    import('./pages/ModerationUsersPage').then(m => ({
        default: m.ModerationUsersPage,
    })),
);
const ChatPage = lazy(() =>
    import('./pages/ChatPage').then(m => ({ default: m.ChatPage })),
);
const ChannelSettingsPage = lazy(() =>
    import('./pages/ChannelSettingsPage').then(m => ({
        default: m.ChannelSettingsPage,
    })),
);
const ForumIndex = lazy(() =>
    import('./pages/forum/ForumIndex').then(m => ({ default: m.ForumIndex })),
);
const ThreadDetail = lazy(() =>
    import('./pages/forum/ThreadDetail').then(m => ({
        default: m.ThreadDetail,
    })),
);
const CreateThread = lazy(() =>
    import('./pages/forum/CreateThread').then(m => ({
        default: m.CreateThread,
    })),
);
const ForumSearchPage = lazy(() =>
    import('./pages/forum/ForumSearchPage').then(m => ({
        default: m.ForumSearchPage,
    })),
);
const ForumAnalyticsPage = lazy(() =>
    import('./pages/forum/ForumAnalyticsPage').then(m => ({
        default: m.ForumAnalyticsPage,
    })),
);
const WebhookSubscriptionsPage = lazy(() =>
    import('./pages/WebhookSubscriptionsPage').then(m => ({
        default: m.WebhookSubscriptionsPage,
    })),
);
const WatchPartyPage = lazy(() =>
    import('./pages/WatchPartyPage').then(m => ({ default: m.WatchPartyPage })),
);
const WatchPartyBrowsePage = lazy(() =>
    import('./pages/WatchPartyBrowsePage').then(m => ({
        default: m.WatchPartyBrowsePage,
    })),
);
const WatchPartyCreatePage = lazy(() =>
    import('./pages/WatchPartyCreatePage').then(m => ({
        default: m.WatchPartyCreatePage,
    })),
);
const WatchPartySettingsPage = lazy(() =>
    import('./pages/WatchPartySettingsPage').then(m => ({
        default: m.WatchPartySettingsPage,
    })),
);

// Loading fallback component
function LoadingFallback() {
    return (
        <div className='min-h-screen flex items-center justify-center'>
            <Spinner size='xl' />
        </div>
    );
}

function App() {
    return (
        <HelmetProvider>
            <AuthProvider>
                <ConsentProvider>
                    <ToastProvider>
                        <BrowserRouter>
                            <Suspense fallback={<LoadingFallback />}>
                                <Routes>
                                    <Route element={<AppLayout />}>
                                        {/* Public Routes */}
                                        <Route
                                            path='/'
                                            element={<HomePage />}
                                        />
                                        <Route
                                            path='/discover'
                                            element={<ScrapedClipsPage />}
                                        />
                                        <Route
                                            path='/discover/lists'
                                            element={<DiscoveryListsPage />}
                                        />
                                        <Route
                                            path='/discover/lists/:id'
                                            element={
                                                <DiscoveryListDetailPage />
                                            }
                                        />
                                        <Route
                                            path='/discover/scraped'
                                            element={<ScrapedClipsPage />}
                                        />
                                        {/* Live Feed - Hidden until after launch */}
                                        {/* <Route
                                            path='/discover/live'
                                            element={
                                                <ProtectedRoute>
                                                    <LiveFeedPage />
                                                </ProtectedRoute>
                                            }
                                        /> */}
                                        <Route
                                            path='/clip/:id'
                                            element={<ClipDetailPage />}
                                        />
                                        <Route
                                            path='/clips/:id'
                                            element={<ClipDetailPage />}
                                        />
                                        <Route
                                            path='/game/:gameId'
                                            element={<GamePage />}
                                        />
                                        <Route
                                            path='/category/:categorySlug'
                                            element={<CategoryPage />}
                                        />
                                        <Route
                                            path='/broadcaster/:broadcasterId'
                                            element={<BroadcasterPage />}
                                        />
                                        <Route
                                            path='/creator/:creatorId'
                                            element={<CreatorPage />}
                                        />
                                        <Route
                                            path='/stream/:streamer'
                                            element={<StreamPage />}
                                        />
                                        <Route
                                            path='/creator/:creatorName/analytics'
                                            element={<CreatorAnalyticsPage />}
                                        />
                                        <Route
                                            path='/user/:username'
                                            element={<UserProfilePage />}
                                        />
                                        <Route
                                            path='/tag/:tagSlug'
                                            element={<TagPage />}
                                        />
                                        <Route
                                            path='/search'
                                            element={<SearchPage />}
                                        />
                                        <Route
                                            path='/about'
                                            element={<AboutPage />}
                                        />
                                        <Route
                                            path='/privacy'
                                            element={<PrivacyPage />}
                                        />
                                        <Route
                                            path='/terms'
                                            element={<TermsPage />}
                                        />
                                        <Route
                                            path='/legal/dmca'
                                            element={<DMCAPage />}
                                        />
                                        <Route
                                            path='/community-rules'
                                            element={<CommunityRulesPage />}
                                        />
                                        <Route
                                            path='/contact'
                                            element={<ContactPage />}
                                        />
                                        <Route
                                            path='/docs'
                                            element={<DocsPage />}
                                        />
                                        <Route
                                            path='/extension'
                                            element={<ExtensionPage />}
                                        />
                                        <Route
                                            path='/leaderboards'
                                            element={<LeaderboardPage />}
                                        />
                                        <Route
                                            path='/pricing'
                                            element={<PricingPage />}
                                        />
                                        <Route
                                            path='/subscription/success'
                                            element={
                                                <SubscriptionSuccessPage />
                                            }
                                        />
                                        <Route
                                            path='/subscription/cancel'
                                            element={<SubscriptionCancelPage />}
                                        />

                                        {/* Forum Routes */}
                                        <Route
                                            path='/forum'
                                            element={<ForumIndex />}
                                        />
                                        <Route
                                            path='/forum/search'
                                            element={<ForumSearchPage />}
                                        />
                                        <Route
                                            path='/forum/analytics'
                                            element={<ForumAnalyticsPage />}
                                        />
                                        <Route
                                            path='/forum/threads/:threadId'
                                            element={<ThreadDetail />}
                                        />
                                        <Route
                                            path='/forum/new'
                                            element={
                                                <ProtectedRoute>
                                                    <CreateThread />
                                                </ProtectedRoute>
                                            }
                                        />

                                        {import.meta.env.DEV && (
                                            <>
                                                <Route
                                                    path='/test/role-badges'
                                                    element={
                                                        <RoleBadgeTestPage />
                                                    }
                                                />
                                                <Route
                                                    path='/test/verified-badge'
                                                    element={
                                                        <VerifiedBadgeTestPage />
                                                    }
                                                />
                                            </>
                                        )}

                                        {/* Guest Routes (redirect to home if authenticated) */}
                                        <Route
                                            path='/login'
                                            element={
                                                <GuestRoute>
                                                    <LoginPage />
                                                </GuestRoute>
                                            }
                                        />

                                        {/* Auth callback route */}
                                        <Route
                                            path='/auth/success'
                                            element={<AuthCallbackPage />}
                                        />

                                        {/* Protected Routes (require authentication) */}
                                        <Route
                                            path='/favorites'
                                            element={
                                                <ProtectedRoute>
                                                    <FavoritesPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/watch-history'
                                            element={
                                                <ProtectedRoute>
                                                    <WatchHistoryPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/queue'
                                            element={
                                                <ProtectedRoute>
                                                    <QueuePage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/queue/theatre'
                                            element={
                                                <ProtectedRoute>
                                                    <QueueTheatrePage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/playlists'
                                            element={
                                                <ProtectedRoute>
                                                    <PlaylistsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/playlists/new'
                                            element={
                                                <ProtectedRoute>
                                                    <PlaylistCreatePage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/playlists/discover'
                                            element={<PublicPlaylistsPage />}
                                        />
                                        <Route
                                            path='/playlists/smart'
                                            element={
                                                <ProtectedRoute>
                                                    <SmartPlaylistsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/playlists/bookmarks'
                                            element={
                                                <ProtectedRoute>
                                                    <BookmarkedPlaylistsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/playlists/:id'
                                            element={<PlaylistDetailPage />}
                                        />
                                        <Route
                                            path='/playlists/:id/theatre'
                                            element={<PlaylistTheatrePage />}
                                        />
                                        <Route
                                            path='/profile'
                                            element={
                                                <ProtectedRoute>
                                                    <ProfilePage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/verification/apply'
                                            element={
                                                <ProtectedRoute>
                                                    <VerificationApplicationPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/settings'
                                            element={
                                                <ProtectedRoute>
                                                    <SettingsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/settings/cookies'
                                            element={
                                                <ProtectedRoute>
                                                    <CookieSettingsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/settings/webhooks'
                                            element={
                                                <ProtectedRoute>
                                                    <WebhookSubscriptionsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/submit'
                                            element={
                                                <ProtectedRoute>
                                                    <SubmitClipPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/submissions'
                                            element={
                                                <ProtectedRoute>
                                                    <UserSubmissionsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/notifications'
                                            element={
                                                <ProtectedRoute>
                                                    <NotificationsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/notifications/preferences'
                                            element={
                                                <ProtectedRoute>
                                                    <NotificationPreferencesPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/profile/stats'
                                            element={
                                                <ProtectedRoute>
                                                    <PersonalStatsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/chat'
                                            element={
                                                <ProtectedRoute>
                                                    <ChatPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/chat/channels/:id/settings'
                                            element={
                                                <ProtectedRoute>
                                                    <ChannelSettingsPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/creator/:creatorId/dashboard'
                                            element={
                                                <ProtectedRoute>
                                                    <CreatorDashboardPage />
                                                </ProtectedRoute>
                                            }
                                        />

                                        {/* Watch Party Routes - Hidden until after launch */}
                                        {/* <Route
                                            path='/watch-parties/browse'
                                            element={<WatchPartyBrowsePage />}
                                        />
                                        <Route
                                            path='/watch-parties/create'
                                            element={
                                                <ProtectedRoute>
                                                    <WatchPartyCreatePage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/watch-parties/:id'
                                            element={
                                                <ProtectedRoute>
                                                    <WatchPartyPage />
                                                </ProtectedRoute>
                                            }
                                        />
                                        <Route
                                            path='/watch-parties/:id/settings'
                                            element={
                                                <ProtectedRoute>
                                                    <WatchPartySettingsPage />
                                                </ProtectedRoute>
                                            }
                                        /> */}

                                        {/* Admin Routes (require admin role) */}
                                        <Route
                                            path='/admin/dashboard'
                                            element={
                                                <AdminRoute>
                                                    <AdminDashboard />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/clips'
                                            element={
                                                <AdminRoute>
                                                    <AdminClipsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/comments'
                                            element={
                                                <AdminRoute>
                                                    <AdminCommentsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/users'
                                            element={
                                                <AdminRoute>
                                                    <AdminUsersPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/reports'
                                            element={
                                                <AdminRoute>
                                                    <AdminReportsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/webhooks/dlq'
                                            element={
                                                <AdminRoute>
                                                    <AdminWebhookDLQPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/sync'
                                            element={
                                                <AdminRoute>
                                                    <AdminSyncPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/analytics'
                                            element={
                                                <AdminRoute>
                                                    <AdminAnalyticsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/revenue'
                                            element={
                                                <AdminRoute>
                                                    <AdminRevenuePage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/campaigns'
                                            element={
                                                <AdminRoute>
                                                    <AdminCampaignsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/submissions'
                                            element={
                                                <AdminRoute>
                                                    <ModerationQueuePage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/moderation'
                                            element={
                                                <AdminRoute>
                                                    <AdminModerationQueuePage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/moderation/analytics'
                                            element={
                                                <AdminRoute>
                                                    <AdminModerationAnalyticsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/moderators'
                                            element={
                                                <AdminRoute>
                                                    <AdminModeratorsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/bans'
                                            element={
                                                <AdminRoute>
                                                    <AdminBansPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/audit-logs'
                                            element={
                                                <AdminRoute>
                                                    <AdminAuditLogsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/verification'
                                            element={
                                                <AdminRoute>
                                                    <AdminVerificationQueuePage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/discovery-lists'
                                            element={
                                                <AdminRoute>
                                                    <AdminDiscoveryListsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/playlist-scripts'
                                            element={
                                                <AdminRoute>
                                                    <AdminPlaylistScriptsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/tags'
                                            element={
                                                <AdminRoute>
                                                    <AdminTagsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/discovery-lists/:id/edit'
                                            element={
                                                <AdminRoute>
                                                    <AdminDiscoveryListFormPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/api-docs'
                                            element={
                                                <AdminRoute>
                                                    <AdminAPIDocsPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/forum/moderation'
                                            element={
                                                <AdminRoute>
                                                    <ForumModerationPage />
                                                </AdminRoute>
                                            }
                                        />
                                        <Route
                                            path='/admin/forum/moderation-log'
                                            element={
                                                <AdminRoute>
                                                    <ModerationLogPage />
                                                </AdminRoute>
                                            }
                                        />

                                        {/* Moderation Routes (require moderator or admin role) */}
                                        <Route
                                            path='/moderation/users'
                                            element={
                                                <AdminRoute>
                                                    <ModerationUsersPage />
                                                </AdminRoute>
                                            }
                                        />

                                        {/* 404 Not Found */}
                                        <Route
                                            path='*'
                                            element={<NotFoundPage />}
                                        />
                                    </Route>
                                </Routes>
                            </Suspense>
                            {/* Consent banner for GDPR compliance */}
                            <ConsentBanner />
                        </BrowserRouter>
                    </ToastProvider>
                </ConsentProvider>
            </AuthProvider>
        </HelmetProvider>
    );
}

export default App;
