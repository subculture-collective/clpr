import type { ElementType } from 'react';
import { Activity, BadgeCheck, Ban, BarChart3, BookOpen, Bot, CircleDollarSign, ClipboardList, FileWarning, Film, Gauge, LayoutDashboard, ListVideo, Megaphone, MessageSquare, Radio, ScrollText, Shield, Tags, Users, Webhook } from 'lucide-react';

export type AdminNavItem = { label: string; href: string; description: string; icon: ElementType };
export type AdminNavGroup = { label: string; items: AdminNavItem[] };

export const adminNavGroups: AdminNavGroup[] = [
    { label: 'Overview', items: [
        { label: 'Control room', href: '/admin/dashboard', description: 'Admin overview and shortcuts', icon: LayoutDashboard },
        { label: 'Platform analytics', href: '/admin/analytics', description: 'Audience and platform health', icon: BarChart3 },
        { label: 'Community support', href: '/admin/revenue', description: 'Open access and Patreon funding', icon: CircleDollarSign },
    ] },
    { label: 'Moderation', items: [
        { label: 'Review queue', href: '/admin/moderation', description: 'Prioritized moderation work', icon: Shield },
        { label: 'Reports', href: '/admin/reports', description: 'User-submitted reports', icon: FileWarning },
        { label: 'Clips', href: '/admin/clips', description: 'Review and manage clips', icon: Film },
        { label: 'Comments', href: '/admin/comments', description: 'Review community replies', icon: MessageSquare },
        { label: 'Submissions', href: '/admin/submissions', description: 'Approve community submissions', icon: ClipboardList },
        { label: 'Verification', href: '/admin/verification', description: 'Creator verification requests', icon: BadgeCheck },
        { label: 'Bans', href: '/admin/bans', description: 'Restrictions and appeals', icon: Ban },
        { label: 'Moderators', href: '/admin/moderators', description: 'Roles and permissions', icon: Users },
        { label: 'User moderation', href: '/moderation/users', description: 'User sanctions and history', icon: Shield },
        { label: 'Moderation analytics', href: '/admin/moderation/analytics', description: 'Workload and outcomes', icon: Gauge },
        { label: 'Audit logs', href: '/admin/audit-logs', description: 'Administrative activity', icon: ScrollText },
    ] },
    { label: 'Discovery', items: [
        { label: 'Curated collections', href: '/admin/discovery-lists', description: 'Featured clip collections', icon: ListVideo },
        { label: 'Playlist automation', href: '/admin/playlist-scripts', description: 'Rules for generated collections', icon: Bot },
        { label: 'Topics', href: '/admin/topics', description: 'Editorial topic taxonomy', icon: ClipboardList },
        { label: 'Tags', href: '/admin/tags', description: 'Tag cleanup and blacklist', icon: Tags },
        { label: 'Tag promotion', href: '/admin/tag-promotion', description: 'Review emerging tags', icon: Activity },
    ] },
    { label: 'Operations', items: [
        { label: 'Sync controls', href: '/admin/sync', description: 'Twitch ingestion controls', icon: Radio },
        { label: 'Webhook failures', href: '/admin/webhooks/dlq', description: 'Replay failed deliveries', icon: Webhook },
        { label: 'Campaigns', href: '/admin/campaigns', description: 'Promotional campaigns', icon: Megaphone },
        { label: 'Users', href: '/admin/users', description: 'Accounts and access', icon: Users },
        { label: 'API reference', href: '/admin/api-docs', description: 'Administrative API docs', icon: BookOpen },
    ] },
];

export const adminNavItems = adminNavGroups.flatMap(group => group.items);
