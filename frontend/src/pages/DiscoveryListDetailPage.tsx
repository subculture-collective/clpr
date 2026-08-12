import { useParams, Navigate } from 'react-router-dom';

// Discovery lists are now playlists - redirect to playlist detail page
export function DiscoveryListDetailPage() {
    const { id } = useParams<{ id: string }>();

    // Redirect to playlist detail page
    const isPlaylistID = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(id || '');
    return <Navigate to={isPlaylistID ? `/playlists/${id}` : '/discover/lists'} replace />;
}
