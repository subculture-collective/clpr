import type { ComponentType } from 'react';
import { Route } from 'react-router-dom';
import { ProtectedRoute } from '@/components/guards/ProtectedRoute';

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

function protectedPage(PageComponent: Page) {
    return (
        <ProtectedRoute>
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
            <Route path='/favorites' element={protectedPage(pages.favorites)} />
            <Route path='/watch-history' element={protectedPage(pages.watchHistory)} />
            <Route path='/queue' element={protectedPage(pages.queue)} />
            <Route path='/queue/theatre' element={protectedPage(pages.queueTheatre)} />
            <Route
                path='/streamer-tools/:channel/clips'
                element={protectedPage(pages.streamerClipRoom)}
            />
            <Route path='/playlists' element={protectedPage(pages.playlists)} />
            <Route path='/playlists/new' element={protectedPage(pages.playlistCreate)} />
            <Route path='/playlists/discover' element={publicPage(pages.publicPlaylists)} />
            <Route path='/playlists/smart' element={protectedPage(pages.smartPlaylists)} />
            <Route
                path='/playlists/bookmarks'
                element={protectedPage(pages.bookmarkedPlaylists)}
            />
            <Route path='/playlists/:id' element={publicPage(pages.playlistDetail)} />
            <Route path='/playlists/:id/theatre' element={publicPage(pages.playlistTheatre)} />
            <Route path='/profile' element={protectedPage(pages.profile)} />
            <Route
                path='/verification/apply'
                element={protectedPage(pages.verificationApplication)}
            />
            <Route path='/settings' element={protectedPage(pages.settings)} />
            <Route path='/settings/cookies' element={publicPage(pages.cookieSettings)} />
            <Route
                path='/settings/webhooks'
                element={protectedPage(pages.webhookSubscriptions)}
            />
            <Route path='/submit' element={protectedPage(pages.submitClip)} />
            <Route path='/submissions' element={protectedPage(pages.submissions)} />
            <Route path='/notifications' element={protectedPage(pages.notifications)} />
            <Route
                path='/notifications/preferences'
                element={protectedPage(pages.notificationPreferences)}
            />
            <Route path='/profile/stats' element={protectedPage(pages.personalStats)} />
            <Route path='/chat' element={protectedPage(pages.chat)} />
            <Route
                path='/chat/channels/:id/settings'
                element={protectedPage(pages.channelSettings)}
            />
            <Route
                path='/creator/:creatorId/dashboard'
                element={protectedPage(pages.creatorDashboard)}
            />
        </>
    );
}
