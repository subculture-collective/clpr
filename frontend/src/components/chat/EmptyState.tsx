import { MessageCircle } from 'lucide-react';

/**
 * EmptyState component shown when no channel is selected
 */
export function EmptyState() {
  return (
    <div className="flex-1 flex items-center justify-center bg-background">
      <div className="text-center max-w-md px-4">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900 mb-4">
          <MessageCircle className="w-8 h-8 text-link dark:text-primary-400" />
        </div>
        <h2 className="text-xl font-semibold text-foreground mb-2">
          Select a channel
        </h2>
        <p className="text-muted-foreground text-sm">
          Choose a channel from the sidebar to start chatting with your community
        </p>
      </div>
    </div>
  );
}
