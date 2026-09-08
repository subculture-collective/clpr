import { ModerationQueueView } from '../../components/moderation/ModerationQueueView';

export function AdminClipsPage() {
    return (
        <ModerationQueueView
            contentType='clip'
            title='Clip Moderation Queue'
            description='Review and moderate flagged clips with bulk actions'
        />
    );
}
