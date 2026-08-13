import type { ComponentType } from 'react';
import { Route } from 'react-router-dom';
import { ProtectedRoute } from '@/components/guards/ProtectedRoute';
import { SEO } from '@/components/SEO';

type Page = ComponentType;

export interface AccountRoutePages {
    favorites: Page;
    watchHistory: Page;
    queue: Page;
    queueTheatre: Page;
    streamerClipRoom: Page;
    playlists: Page;
    playlistCreate: Page;
    publicPlaylists: Page;
    smartPlaylists: Page;
    bookmarkedPlaylists: Page;
    playlistDetail: Page;
    playlistTheatre: Page;
    profile: Page;
    verificationApplication: Page;
    settings: Page;
    cookieSettings: Page;
    webhookSubscriptions: Page;
    submitClip: Page;
    submissions: Page;
    notifications: Page;
    notificationPreferences: Page;
    personalStats: Page;
    chat: Page;
    channelSettings: Page;
    creatorDashboard: Page;
}

function protectedPage(PageComponent: Page, title: string) {
    return (
        <ProtectedRoute>
            <SEO title={title} noindex />
            <PageComponent />
        </ProtectedRoute>
    );
}

function publicPage(PageComponent: Page) {
    return <PageComponent />;
}

/** Versioned signed-in product routes. Public playlist views remain explicitly unwrapped. */
export function accountRoutes(pages: AccountRoutePages) {
    return (
        <>
            <Route path='/favorites' element={protectedPage(pages.favorites, 'Favorites')} />
            <Route path='/watch-history' element={protectedPage(pages.watchHistory, 'Watch History')} />
            <Route path='/queue' element={protectedPage(pages.queue, 'Queue')} />
            <Route path='/queue/theatre' element={protectedPage(pages.queueTheatre, 'Queue Theatre')} />
            <Route
                path='/streamer-tools/:channel/clips'
                element={protectedPage(pages.streamerClipRoom, 'Streamer Clip Room')}
            />
            <Route path='/playlists' element={protectedPage(pages.playlists, 'My Playlists')} />
            <Route path='/playlists/new' element={protectedPage(pages.playlistCreate, 'Create Playlist')} />
            <Route path='/playlists/discover' element={publicPage(pages.publicPlaylists)} />
            <Route path='/playlists/smart' element={protectedPage(pages.smartPlaylists, 'Smart Playlists')} />
            <Route
                path='/playlists/bookmarks'
                element={protectedPage(pages.bookmarkedPlaylists, 'Bookmarked Playlists')}
            />
            <Route path='/playlists/:id' element={publicPage(pages.playlistDetail)} />
            <Route path='/playlists/:id/theatre' element={publicPage(pages.playlistTheatre)} />
            <Route path='/profile' element={protectedPage(pages.profile, 'Profile')} />
            <Route
                path='/verification/apply'
                element={protectedPage(pages.verificationApplication, 'Creator Verification')}
            />
            <Route path='/settings' element={protectedPage(pages.settings, 'Settings')} />
            <Route path='/settings/cookies' element={publicPage(pages.cookieSettings)} />
            <Route
                path='/settings/webhooks'
                element={protectedPage(pages.webhookSubscriptions, 'Webhook Subscriptions')}
            />
            <Route path='/submit' element={protectedPage(pages.submitClip, 'Submit a Clip')} />
            <Route path='/submissions' element={protectedPage(pages.submissions, 'My Submissions')} />
            <Route path='/notifications' element={protectedPage(pages.notifications, 'Notifications')} />
            <Route
                path='/notifications/preferences'
                element={protectedPage(pages.notificationPreferences, 'Notification Preferences')}
            />
            <Route path='/profile/stats' element={protectedPage(pages.personalStats, 'Personal Stats')} />
            <Route path='/chat' element={protectedPage(pages.chat, 'Chat')} />
            <Route
                path='/chat/channels/:id/settings'
                element={protectedPage(pages.channelSettings, 'Channel Settings')}
            />
            <Route
                path='/creator/:creatorId/dashboard'
                element={protectedPage(pages.creatorDashboard, 'Creator Dashboard')}
            />
        </>
    );
}
