import { ModerationQueueView } from '../../components/moderation/ModerationQueueView';

export function AdminCommentsPage() {
    return (
        <ModerationQueueView
            contentType='comment'
            title='Comment Moderation Queue'
            description='Review and moderate flagged comments with bulk actions'
        />
    );
}
