import { useState } from 'react';
import { Search, Hash } from 'lucide-react';
import type { Channel } from '@/types/chat';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/utils';

interface ChannelSidebarProps {
  channels: Channel[];
  selectedChannel: string | null;
  onSelectChannel: (channelId: string) => void;
}

/**
 * ChannelSidebar component with search and channel list
 */
export function ChannelSidebar({
  channels,
  selectedChannel,
  onSelectChannel,
}: ChannelSidebarProps) {
  const [searchQuery, setSearchQuery] = useState('');

  const filteredChannels = channels.filter((channel) =>
    channel.name.toLocaleLowerCase().includes(searchQuery.toLocaleLowerCase())
  );

  return (
    <div className="w-full md:w-64 bg-neutral-900 dark:bg-neutral-950 text-white flex flex-col border-r border-neutral-800">
      {/* Header */}
      <div className="p-4 border-b border-neutral-800">
        <h2 className="text-lg font-semibold mb-3">Channels</h2>
        <Input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Search channels..."
          leftIcon={<Search className="w-4 h-4" />}
          className="bg-neutral-800 border-neutral-700 text-white placeholder:text-neutral-400"
        />
      </div>

      {/* Channel list */}
      <div className="flex-1 overflow-y-auto">
        {filteredChannels.length === 0 ? (
          <div className="p-4 text-center text-neutral-400 text-sm">
            {searchQuery ? 'No channels found' : 'No channels available'}
          </div>
        ) : (
          <div className="py-2">
            {filteredChannels.map((channel) => (
              <button
                key={channel.id}
                onClick={() => onSelectChannel(channel.id)}
                className={cn(
                  'w-full text-left px-4 py-2.5 hover:bg-neutral-800 transition-colors flex items-center justify-between gap-2 group',
                  selectedChannel === channel.id && 'bg-neutral-800 border-l-2 border-primary-500'
                )}
                aria-current={selectedChannel === channel.id ? 'page' : undefined}
              >
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  <Hash className="w-4 h-4 text-neutral-400 flex-shrink-0" />
                  <span className="truncate font-medium">{channel.name}</span>
                </div>
                
                {channel.unread_count && channel.unread_count > 0 && (
                  <span
                    className="bg-error-500 text-white text-xs px-2 py-0.5 font-medium flex-shrink-0"
                    aria-label={`${channel.unread_count} unread messages`}
                  >
                    {channel.unread_count > 99 ? '99+' : channel.unread_count}
                  </span>
                )}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Footer info */}
      <div className="p-3 border-t border-neutral-800 text-xs text-neutral-400">
        {filteredChannels.length} {filteredChannels.length === 1 ? 'channel' : 'channels'}
      </div>
    </div>
  );
}
