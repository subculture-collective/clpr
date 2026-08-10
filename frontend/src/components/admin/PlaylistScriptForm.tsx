import { useState, useCallback, useRef } from 'react';
import { ChevronDown, ChevronRight, Info, Plus, X } from 'lucide-react';
import {
    Card,
    CardHeader,
    CardBody,
    Button,
    Input,
    TextArea,
    Toggle,
    Badge,
} from '@/components';
import { cn } from '@/lib/utils';
import {
    STRATEGY_META,
    SCHEDULE_LABELS,
    TITLE_PLACEHOLDERS,
    buildPlaylistTitle,
    buildScriptSummary,
} from '@/lib/playlist-script-utils';
import type {
    PlaylistScriptSort,
    PlaylistScriptTimeframe,
    PlaylistScriptSchedule,
    PlaylistScriptStrategy,
    PlaylistScriptVisibility,
} from '@/types/playlistScript';

export interface PlaylistScriptFormValues {
    name: string;
    description: string;
    visibility: PlaylistScriptVisibility;
    is_active: boolean;
    strategy: PlaylistScriptStrategy;
    sort: PlaylistScriptSort;
    timeframe: PlaylistScriptTimeframe;
    clip_limit: number;
    game_id: string;
    game_ids: string[];
    broadcaster_id: string;
    tag: string;
    tags: string[];
    tags_logic: 'and' | 'or';
    exclude_tags: string[];
    language: string;
    min_vote_score: string;
    min_view_count: string;
    exclude_nsfw: boolean;
    top_10k_streamers: boolean;
    seed_clip_id: string;
    schedule: PlaylistScriptSchedule;
    retention_days: number;
    title_template: string;
}

const DEFAULT_VALUES: PlaylistScriptFormValues = {
    name: '',
    description: '',
    visibility: 'public',
    is_active: true,
    strategy: 'standard',
    sort: 'top',
    timeframe: 'day',
    clip_limit: 10,
    game_id: '',
    game_ids: [],
    broadcaster_id: '',
    tag: '',
    tags: [],
    tags_logic: 'and',
    exclude_tags: [],
    language: '',
    min_vote_score: '',
    min_view_count: '',
    exclude_nsfw: true,
    top_10k_streamers: false,
    seed_clip_id: '',
    schedule: 'daily',
    retention_days: 30,
    title_template: '',
};

interface Props {
    initialValues?: PlaylistScriptFormValues;
    onSubmit: (values: PlaylistScriptFormValues) => void;
    onCancel?: () => void;
    isEditing?: boolean;
    isLoading?: boolean;
    /** Limit strategy to 'standard' only (for user-facing smart playlists) */
    standardOnly?: boolean;
}

function CollapsibleSection({
    title,
    defaultOpen = true,
    children,
}: {
    title: string;
    defaultOpen?: boolean;
    children: React.ReactNode;
}) {
    const [open, setOpen] = useState(defaultOpen);
    return (
        <div className="border border-border rounded-lg">
            <button
                type="button"
                onClick={() => setOpen((o) => !o)}
                className="flex items-center justify-between w-full px-4 py-3 text-left font-medium hover:bg-accent/50 transition-colors rounded-lg"
            >
                {title}
                {open ? (
                    <ChevronDown className="w-4 h-4" />
                ) : (
                    <ChevronRight className="w-4 h-4" />
                )}
            </button>
            {open && <div className="px-4 pb-4">{children}</div>}
        </div>
    );
}

function TagInput({
    value,
    onChange,
    placeholder,
}: {
    value: string[];
    onChange: (v: string[]) => void;
    placeholder?: string;
}) {
    const [input, setInput] = useState('');

    const addTag = () => {
        const trimmed = input.trim();
        if (trimmed && !value.includes(trimmed)) {
            onChange([...value, trimmed]);
        }
        setInput('');
    };

    return (
        <div>
            <div className="flex flex-wrap gap-1.5 mb-2">
                {value.map((tag) => (
                    <Badge key={tag} variant="default" size="sm">
                        {tag}
                        <button
                            type="button"
                            onClick={() =>
                                onChange(value.filter((t) => t !== tag))
                            }
                            className="ml-1 hover:text-error-500"
                        >
                            <X className="w-3 h-3" />
                        </button>
                    </Badge>
                ))}
            </div>
            <div className="flex gap-2">
                <input
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                            e.preventDefault();
                            addTag();
                        }
                    }}
                    placeholder={placeholder}
                    className="flex-1 px-3 py-2 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                />
                <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={addTag}
                >
                    <Plus className="w-4 h-4" />
                </Button>
            </div>
        </div>
    );
}

export function PlaylistScriptForm({
    initialValues,
    onSubmit,
    onCancel,
    isEditing = false,
    isLoading = false,
    standardOnly = false,
}: Props) {
    const [form, setForm] = useState<PlaylistScriptFormValues>(
        initialValues || { ...DEFAULT_VALUES },
    );
    const titleTemplateRef = useRef<HTMLInputElement>(null);

    const set = useCallback(
        <K extends keyof PlaylistScriptFormValues>(
            key: K,
            value: PlaylistScriptFormValues[K],
        ) => {
            setForm((prev) => ({ ...prev, [key]: value }));
        },
        [],
    );

    const strategyMeta = STRATEGY_META[form.strategy];
    const showSortTimeframe = form.strategy === 'standard';

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(form);
    };

    const insertPlaceholder = (token: string) => {
        const input = titleTemplateRef.current;
        if (!input) {
            set('title_template', form.title_template + token);
            return;
        }
        const start = input.selectionStart ?? form.title_template.length;
        const end = input.selectionEnd ?? start;
        const next =
            form.title_template.slice(0, start) +
            token +
            form.title_template.slice(end);
        set('title_template', next);
        // Restore cursor after React re-render
        requestAnimationFrame(() => {
            input.focus();
            const pos = start + token.length;
            input.setSelectionRange(pos, pos);
        });
    };

    const titlePreview = buildPlaylistTitle({
        name: form.name || 'Untitled',
        title_template: form.title_template || undefined,
    });

    const summary = buildScriptSummary(form);

    const databaseStrategies = Object.entries(STRATEGY_META).filter(
        ([, m]) => m.group === 'database',
    );
    const twitchStrategies = Object.entries(STRATEGY_META).filter(
        ([, m]) => m.group === 'twitch',
    );

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            {/* Section 1 — Basics */}
            <Card>
                <CardHeader>
                    <h3 className="text-lg font-semibold">
                        {isEditing ? 'Edit Script' : 'Create Script'}
                    </h3>
                </CardHeader>
                <CardBody className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <Input
                            label="Name"
                            value={form.name}
                            onChange={(e) => set('name', e.target.value)}
                            placeholder="Daily Top 10"
                            maxLength={100}
                            required
                            fullWidth
                        />
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Visibility
                            </label>
                            <select
                                value={form.visibility}
                                onChange={(e) =>
                                    set(
                                        'visibility',
                                        e.target
                                            .value as PlaylistScriptVisibility,
                                    )
                                }
                                className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                            >
                                <option value="public">Public</option>
                                <option value="unlisted">Unlisted</option>
                                <option value="private">Private</option>
                            </select>
                        </div>
                    </div>
                    <TextArea
                        label="Description"
                        value={form.description}
                        onChange={(e) => set('description', e.target.value)}
                        placeholder="Auto-generated playlist of top clips"
                        maxLength={500}
                        showCount
                        fullWidth
                        rows={2}
                    />
                    <Toggle
                        label="Active"
                        helperText="When active, scheduled scripts run automatically"
                        checked={form.is_active}
                        onChange={(e) => set('is_active', e.target.checked)}
                    />
                </CardBody>
            </Card>

            {/* Section 2 — Strategy */}
            <Card>
                <CardHeader>
                    <h3 className="text-lg font-semibold">Strategy</h3>
                </CardHeader>
                <CardBody className="space-y-4">
                    {!standardOnly && (
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Curation Strategy
                            </label>
                            <select
                                value={form.strategy}
                                onChange={(e) =>
                                    set(
                                        'strategy',
                                        e.target
                                            .value as PlaylistScriptStrategy,
                                    )
                                }
                                className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                            >
                                <optgroup label="Database Strategies">
                                    {databaseStrategies.map(([key, meta]) => (
                                        <option key={key} value={key}>
                                            {meta.label} — {meta.description}
                                        </option>
                                    ))}
                                </optgroup>
                                <optgroup label="Twitch-Powered">
                                    {twitchStrategies.map(([key, meta]) => (
                                        <option key={key} value={key}>
                                            {meta.label} — {meta.description}
                                        </option>
                                    ))}
                                </optgroup>
                            </select>
                            {strategyMeta && (
                                <p className="text-sm text-muted-foreground flex items-start gap-1.5">
                                    <Info className="w-4 h-4 mt-0.5 shrink-0" />
                                    {strategyMeta.description}
                                </p>
                            )}
                        </div>
                    )}

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {showSortTimeframe && (
                            <>
                                <div className="flex flex-col gap-1.5">
                                    <label className="text-sm font-medium text-foreground">
                                        Sort
                                    </label>
                                    <select
                                        value={form.sort}
                                        onChange={(e) =>
                                            set(
                                                'sort',
                                                e.target
                                                    .value as PlaylistScriptSort,
                                            )
                                        }
                                        className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                                    >
                                        <option value="top">Top</option>
                                        <option value="trending">
                                            Trending
                                        </option>
                                        <option value="hot">Hot</option>
                                        <option value="popular">Popular</option>
                                        <option value="new">New</option>
                                        <option value="rising">Rising</option>
                                        <option value="discussed">
                                            Discussed
                                        </option>
                                    </select>
                                </div>
                                <div className="flex flex-col gap-1.5">
                                    <label className="text-sm font-medium text-foreground">
                                        Timeframe
                                    </label>
                                    <select
                                        value={form.timeframe}
                                        onChange={(e) =>
                                            set(
                                                'timeframe',
                                                e.target
                                                    .value as PlaylistScriptTimeframe,
                                            )
                                        }
                                        className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                                    >
                                        <option value="hour">Last hour</option>
                                        <option value="day">Last day</option>
                                        <option value="week">Last week</option>
                                        <option value="month">
                                            Last month
                                        </option>
                                        <option value="year">Last year</option>
                                    </select>
                                </div>
                            </>
                        )}
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Clip Limit
                            </label>
                            <input
                                type="number"
                                min={1}
                                max={100}
                                value={form.clip_limit}
                                onChange={(e) =>
                                    set('clip_limit', Number(e.target.value))
                                }
                                className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                            />
                        </div>
                    </div>

                    {/* Strategy-specific fields */}
                    {strategyMeta?.requiresSeedClipId && (
                        <Input
                            label="Seed Clip ID"
                            value={form.seed_clip_id}
                            onChange={(e) =>
                                set('seed_clip_id', e.target.value)
                            }
                            placeholder="UUID of the seed clip"
                            helperText="Clips similar to this one will be selected"
                            fullWidth
                        />
                    )}
                    {strategyMeta?.requiresGameId && (
                        <Input
                            label="Game ID"
                            value={form.game_id}
                            onChange={(e) => set('game_id', e.target.value)}
                            placeholder="Twitch game ID"
                            fullWidth
                        />
                    )}
                    {strategyMeta?.requiresGameIds && (
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Game IDs
                            </label>
                            <TagInput
                                value={form.game_ids}
                                onChange={(v) => set('game_ids', v)}
                                placeholder="Add a Twitch game ID and press Enter"
                            />
                        </div>
                    )}
                    {strategyMeta?.requiresBroadcasterId && (
                        <Input
                            label="Broadcaster ID"
                            value={form.broadcaster_id}
                            onChange={(e) =>
                                set('broadcaster_id', e.target.value)
                            }
                            placeholder="Twitch broadcaster ID"
                            fullWidth
                        />
                    )}
                </CardBody>
            </Card>

            {/* Section 3 — Advanced Filters */}
            {!standardOnly && (
                <CollapsibleSection
                    title="Advanced Filters"
                    defaultOpen={false}
                >
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {!strategyMeta?.requiresGameId && (
                            <Input
                                label="Game ID"
                                value={form.game_id}
                                onChange={(e) => set('game_id', e.target.value)}
                                placeholder="Filter by Twitch game ID"
                                helperText="Optional single-game filter"
                                fullWidth
                            />
                        )}
                        {!strategyMeta?.requiresBroadcasterId && (
                            <Input
                                label="Broadcaster ID"
                                value={form.broadcaster_id}
                                onChange={(e) =>
                                    set('broadcaster_id', e.target.value)
                                }
                                placeholder="Filter by broadcaster ID"
                                fullWidth
                            />
                        )}
<Input
                            label="Tag (deprecated)"
                            value={form.tag}
                            onChange={(e) => set('tag', e.target.value)}
                            placeholder="Include clips with this tag"
                            helperText="Use the Tags field below for multiple tags with AND/OR logic"
                            fullWidth
                        />
                        {/* Multi-tag filter — full width */}
                        <div className="md:col-span-2 space-y-3">
                            <div className="flex flex-col gap-1.5">
                                <label className="text-sm font-medium text-foreground">
                                    Tags
                                </label>
                                <TagInput
                                    value={form.tags}
                                    onChange={(v) => set('tags', v)}
                                    placeholder="Add a tag slug and press Enter"
                                />
                                <p className="text-sm text-muted-foreground">
                                    Filter clips by tag slugs (e.g.
                                    content/clutch, game/valorant)
                                </p>
                            </div>
                            {form.tags.length > 1 && (
                                <div className="flex flex-col gap-1.5">
                                    <label className="text-sm font-medium text-foreground">
                                        Tags Logic
                                    </label>
                                    <select
                                        value={form.tags_logic}
                                        onChange={(e) =>
                                            set(
                                                'tags_logic',
                                                e.target.value as
                                                    | 'and'
                                                    | 'or',
                                            )
                                        }
                                        className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                                    >
                                        <option value="and">
                                            Must have ALL tags (AND)
                                        </option>
                                        <option value="or">
                                            Must have ANY tag (OR)
                                        </option>
                                    </select>
                                </div>
                            )}
                        </div>
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Exclude Tags
                            </label>
                            <TagInput
                                value={form.exclude_tags}
                                onChange={(v) => set('exclude_tags', v)}
                                placeholder="Tag to exclude"
                            />
                        </div>
                        <Input
                            label="Language"
                            value={form.language}
                            onChange={(e) => set('language', e.target.value)}
                            placeholder="e.g. en"
                            helperText="ISO 639-1 language code"
                            fullWidth
                        />
                        <Input
                            label="Min Vote Score"
                            type="number"
                            min={0}
                            value={form.min_vote_score}
                            onChange={(e) =>
                                set('min_vote_score', e.target.value)
                            }
                            placeholder="0"
                            fullWidth
                        />
                        <Input
                            label="Min View Count"
                            type="number"
                            min={0}
                            value={form.min_view_count}
                            onChange={(e) =>
                                set('min_view_count', e.target.value)
                            }
                            placeholder="0"
                            fullWidth
                        />
                    </div>
                    <div className="mt-4 flex flex-wrap gap-6">
                        <Toggle
                            label="Exclude NSFW"
                            checked={form.exclude_nsfw}
                            onChange={(e) =>
                                set('exclude_nsfw', e.target.checked)
                            }
                        />
                        <Toggle
                            label="Top 10K Streamers Only"
                            checked={form.top_10k_streamers}
                            onChange={(e) =>
                                set('top_10k_streamers', e.target.checked)
                            }
                        />
                    </div>
                </CollapsibleSection>
            )}

            {/* Section 4 — Schedule & Lifecycle */}
            <Card>
                <CardHeader>
                    <h3 className="text-lg font-semibold">
                        Schedule & Lifecycle
                    </h3>
                </CardHeader>
                <CardBody className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Schedule
                            </label>
                            <select
                                value={form.schedule}
                                onChange={(e) =>
                                    set(
                                        'schedule',
                                        e.target
                                            .value as PlaylistScriptSchedule,
                                    )
                                }
                                className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                            >
                                {(standardOnly
                                    ? ([
                                          'manual',
                                          'daily',
                                          'weekly',
                                      ] as PlaylistScriptSchedule[])
                                    : (Object.keys(
                                          SCHEDULE_LABELS,
                                      ) as PlaylistScriptSchedule[])
                                ).map((key) => (
                                    <option key={key} value={key}>
                                        {SCHEDULE_LABELS[key]}
                                    </option>
                                ))}
                            </select>
                            <p className="text-sm text-muted-foreground">
                                {form.schedule === 'manual'
                                    ? 'Manual scripts never run automatically, even when Active is enabled. Use Generate or choose a schedule.'
                                    : !form.is_active
                                      ? 'Inactive scripts will not run until Active is turned back on.'
                                      : 'Scheduled scripts are picked up automatically by the background playlist scheduler.'}
                            </p>
                        </div>
                        <div className="flex flex-col gap-1.5">
                            <label className="text-sm font-medium text-foreground">
                                Retention (days)
                            </label>
                            <input
                                type="number"
                                min={1}
                                max={365}
                                value={form.retention_days}
                                onChange={(e) =>
                                    set(
                                        'retention_days',
                                        Number(e.target.value),
                                    )
                                }
                                className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                            />
                            <p className="text-sm text-muted-foreground">
                                Generated playlists older than{' '}
                                {form.retention_days} day
                                {form.retention_days !== 1 ? 's' : ''} are
                                auto-deleted
                            </p>
                        </div>
                    </div>

                    {/* Title Template */}
                    <div className="flex flex-col gap-1.5">
                        <label className="text-sm font-medium text-foreground">
                            Title Template
                        </label>
                        <input
                            ref={titleTemplateRef}
                            value={form.title_template}
                            onChange={(e) =>
                                set('title_template', e.target.value)
                            }
                            placeholder="{name} • {date}"
                            maxLength={200}
                            className="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground min-h-[44px] focus:outline-none focus:ring-2 focus:ring-primary-500"
                        />
                        <div className="flex flex-wrap gap-1.5">
                            {TITLE_PLACEHOLDERS.map(({ token, label }) => (
                                <button
                                    key={token}
                                    type="button"
                                    onClick={() => insertPlaceholder(token)}
                                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-mono bg-accent text-accent-foreground hover:bg-accent/80 transition-colors"
                                    title={label}
                                >
                                    {token}
                                </button>
                            ))}
                        </div>
                        <div className="mt-1 px-3 py-2 bg-accent/50 rounded-lg text-sm">
                            <span className="text-muted-foreground">
                                Preview:{' '}
                            </span>
                            <span className="font-medium">{titlePreview}</span>
                        </div>
                    </div>
                </CardBody>
            </Card>

            {/* Summary */}
            <div
                className={cn(
                    'px-4 py-3 rounded-lg text-sm',
                    'bg-info-100 text-info-800 dark:bg-info-900 dark:text-info-100',
                )}
            >
                {summary}
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-3">
                {onCancel && (
                    <Button
                        type="button"
                        variant="ghost"
                        onClick={onCancel}
                        disabled={isLoading}
                    >
                        Cancel
                    </Button>
                )}
                <Button
                    type="submit"
                    loading={isLoading}
                    disabled={!form.name.trim()}
                >
                    {isEditing ? 'Update Script' : 'Create Script'}
                </Button>
            </div>
        </form>
    );
}
