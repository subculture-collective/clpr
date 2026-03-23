import React from 'react';
import { cn } from '@/lib/utils';

interface TabItem {
  id: string;
  label: string;
  badge?: number;
}

interface TabProps {
  tabs: TabItem[];
  activeTab: string;
  onTabChange: (tabId: string) => void;
  className?: string;
}

export const Tab: React.FC<TabProps> = ({ tabs, activeTab, onTabChange, className }) => {
  return (
    <div className={cn('flex bg-surface border-b border-border', className)}>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onTabChange(tab.id)}
          className={cn(
            'flex items-center gap-1.5 px-4 py-2.5 text-[13px] font-semibold transition-colors cursor-pointer',
            'font-heading',
            activeTab === tab.id
              ? 'text-text-primary border-b-2 border-brand'
              : 'text-text-secondary hover:text-text-primary'
          )}
        >
          {tab.label}
          {tab.badge !== undefined && tab.badge > 0 && (
            <span className="px-1.5 py-0.5 text-[11px] font-semibold bg-brand text-white rounded-full min-w-[20px] text-center">
              {tab.badge > 99 ? '99+' : tab.badge}
            </span>
          )}
        </button>
      ))}
    </div>
  );
};
