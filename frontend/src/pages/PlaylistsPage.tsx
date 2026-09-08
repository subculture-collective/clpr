import { Container, ScrollToTop } from '../components';
import { MiniFooter } from '../components/layout';
import { PlaylistManager } from '../components/playlist';
import { SEO } from '../components';

export function PlaylistsPage() {
    return (
        <>
            <SEO
                title="My Playlists"
                description="Manage your playlists and organize your favorite clips"
            />
            <Container>
                <div className="py-8">
                    <h1 className="mb-6 text-2xl font-semibold">Your playlists</h1>
                    <PlaylistManager />
                </div>
                <MiniFooter />
            </Container>
            <ScrollToTop />
        </>
    );
}
