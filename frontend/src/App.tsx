import { lazy, Suspense } from 'react';
import { BrowserRouter, Navigate, Routes, Route, useLocation, useParams } from 'react-router-dom';
import { HelmetProvider } from '@dr.pogodin/react-helmet';
import { AuthProvider } from './context/AuthContext';
import { ToastProvider } from './context/ToastContext';
import { ConsentProvider } from './context/ConsentContext';
import { AppLayout } from './components/layout/AppLayout';
import { ProtectedRoute } from './components/guards/ProtectedRoute';
import { GuestRoute } from './components/guards/GuestRoute';
import { Spinner } from './components/ui/Spinner';
import { ConsentBanner } from './components/consent/ConsentBanner';
import { adminRoutes } from './routes/v1/AdminRoutes';
import { accountRoutes } from './routes/v1/AccountRoutes';

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
const ClipDetailPage = lazy(() =>
    import('./pages/ClipDetailPage').then(m => ({ default: m.ClipDetailPage })),
);
const GamePage = lazy(() =>
    import('./pages/GamePage').then(m => ({ default: m.GamePage })),
);

function LegacyGameRedirect() {
    const { gameId } = useParams<{ gameId: string }>();
    const location = useLocation();
    return <Navigate to={`/twitch-category/${gameId || ''}${location.search}`} replace />;
}
const CategoryPage = lazy(() =>
    import('./pages/CategoryPage').then(m => ({ default: m.CategoryPage })),
);
const TopicsPage = lazy(() =>
    import('./pages/TopicsPage').then(m => ({ default: m.TopicsPage })),
);
const BroadcasterPage = lazy(() =>
    import('./pages/BroadcasterPage').then(m => ({
        default: m.BroadcasterPage,
    })),
);
const CreatorsPage = lazy(() =>
    import('./pages/CreatorsPage').then(m => ({ default: m.CreatorsPage })),
);
const OnboardingPage = lazy(() =>
    import('./pages/OnboardingPage').then(m => ({ default: m.OnboardingPage })),
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
const TagsPage = lazy(() =>
    import('./pages/TagsPage').then(m => ({ default: m.TagsPage })),
);

function LegacyTopicRedirect() {
    const { categorySlug } = useParams<{ categorySlug: string }>();
    const location = useLocation();
    return <Navigate to={`/topics/${categorySlug || ''}${location.search}`} replace />;
}

function LegacyTagRedirect() {
    const params = useParams<{ '*': string }>();
    const location = useLocation();
    return <Navigate to={`/tags/${params['*'] || ''}${location.search}`} replace />;
}
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
const AdminTopicsPage = lazy(() =>
    import('./pages/admin/AdminTopicsPage').then(m => ({
        default: m.AdminTopicsPage,
    })),
);
const AdminTagPromotionPage = lazy(() =>
    import('./pages/admin/AdminTagPromotionPage').then(m => ({
        default: m.AdminTagPromotionPage,
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
const StreamerClipRoomPage = lazy(() =>
    import('./pages/StreamerClipRoomPage').then(m => ({
        default: m.StreamerClipRoomPage,
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
                                            element={<Navigate to='/' replace />}
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
                                            element={<Navigate to='/' replace />}
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
                                            element={<LegacyGameRedirect />}
                                        />
                                        <Route
                                            path='/twitch-category/:gameId'
                                            element={<GamePage />}
                                        />
                                        <Route
                                            path='/category/:categorySlug'
                                            element={<LegacyTopicRedirect />}
                                        />
                                        <Route path='/categories' element={<Navigate to='/topics' replace />} />
                                        <Route path='/topics' element={<TopicsPage />} />
                                        <Route path='/topics/:categorySlug' element={<CategoryPage />} />
                                        <Route
                                            path='/broadcaster/:broadcasterId'
                                            element={<BroadcasterPage />}
                                        />
                                        <Route
                                            path='/creators'
                                            element={<CreatorsPage />}
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
                                            path='/tag/*'
                                            element={<LegacyTagRedirect />}
                                        />
                                        <Route path='/tags' element={<TagsPage />} />
                                        <Route path='/tags/:tagSlug' element={<TagPage />} />
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
                                        <Route path='/onboarding' element={<ProtectedRoute><OnboardingPage /></ProtectedRoute>} />

                                        {accountRoutes({
                                            favorites: FavoritesPage,
                                            watchHistory: WatchHistoryPage,
                                            queue: QueuePage,
                                            queueTheatre: QueueTheatrePage,
                                            streamerClipRoom: StreamerClipRoomPage,
                                            playlists: PlaylistsPage,
                                            playlistCreate: PlaylistCreatePage,
                                            publicPlaylists: PublicPlaylistsPage,
                                            smartPlaylists: SmartPlaylistsPage,
                                            bookmarkedPlaylists: BookmarkedPlaylistsPage,
                                            playlistDetail: PlaylistDetailPage,
                                            playlistTheatre: PlaylistTheatrePage,
                                            profile: ProfilePage,
                                            verificationApplication: VerificationApplicationPage,
                                            settings: SettingsPage,
                                            cookieSettings: CookieSettingsPage,
                                            webhookSubscriptions: WebhookSubscriptionsPage,
                                            submitClip: SubmitClipPage,
                                            submissions: UserSubmissionsPage,
                                            notifications: NotificationsPage,
                                            notificationPreferences: NotificationPreferencesPage,
                                            personalStats: PersonalStatsPage,
                                            chat: ChatPage,
                                            channelSettings: ChannelSettingsPage,
                                            creatorDashboard: CreatorDashboardPage,
                                        })}

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

                                        {adminRoutes({
                                                dashboard: AdminDashboard,
                                                clips: AdminClipsPage,
                                                comments: AdminCommentsPage,
                                                users: AdminUsersPage,
                                                reports: AdminReportsPage,
                                                webhookDlq: AdminWebhookDLQPage,
                                                sync: AdminSyncPage,
                                                analytics: AdminAnalyticsPage,
                                                revenue: AdminRevenuePage,
                                                campaigns: AdminCampaignsPage,
                                                submissions: ModerationQueuePage,
                                                moderation: AdminModerationQueuePage,
                                                moderationAnalytics: AdminModerationAnalyticsPage,
                                                moderators: AdminModeratorsPage,
                                                bans: AdminBansPage,
                                                auditLogs: AdminAuditLogsPage,
                                                verification: AdminVerificationQueuePage,
                                                discoveryLists: AdminDiscoveryListsPage,
                                                discoveryListForm: AdminDiscoveryListFormPage,
                                                playlistScripts: AdminPlaylistScriptsPage,
                                                tags: AdminTagsPage,
                                                topics: AdminTopicsPage,
                                                tagPromotion: AdminTagPromotionPage,
                                                apiDocs: AdminAPIDocsPage,
                                                forumModeration: ForumModerationPage,
                                                forumModerationLog: ModerationLogPage,
                                                moderationUsers: ModerationUsersPage,
                                            })}

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
