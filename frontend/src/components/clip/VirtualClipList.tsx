import { useCallback, useLayoutEffect, useRef, useState, type ReactNode } from 'react';
import { defaultRangeExtractor, useWindowVirtualizer, type Range } from '@tanstack/react-virtual';
import type { Clip } from '@/types/clip';

interface VirtualClipListProps {
    clips: Clip[];
    children: (clip: Clip, index: number) => ReactNode;
}

/** Measure rows in both directions; retain keyboard focus outside the overscan window. */
export function VirtualClipList({ clips, children }: VirtualClipListProps) {
    const listRef = useRef<HTMLDivElement>(null);
    const [scrollMargin, setScrollMargin] = useState(0);
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null);
    const getItemKey = useCallback((index: number) => clips[index].id, [clips]);
    const rangeExtractor = useCallback((range: Range) => {
        const indices = defaultRangeExtractor(range);
        if (focusedIndex !== null && focusedIndex < clips.length && !indices.includes(focusedIndex)) {
            indices.push(focusedIndex);
            indices.sort((a, b) => a - b);
        }
        return indices;
    }, [focusedIndex, clips.length]);
    const virtualizer = useWindowVirtualizer({
        count: clips.length,
        getItemKey,
        estimateSize: () => 720,
        overscan: 3,
        scrollMargin,
        rangeExtractor,
        initialRect: { width: window.innerWidth, height: window.innerHeight },
    });

    useLayoutEffect(() => {
        const updateOffset = () => {
            if (listRef.current) setScrollMargin(listRef.current.getBoundingClientRect().top + window.scrollY);
        };
        updateOffset();
        const observer = new ResizeObserver(updateOffset);
        observer.observe(document.body);
        window.addEventListener('resize', updateOffset);
        return () => {
            observer.disconnect();
            window.removeEventListener('resize', updateOffset);
        };
    }, []);

    return (
        <div
            ref={listRef}
            role='list'
            aria-label='Clips'
            className='relative w-full'
            style={{ height: virtualizer.getTotalSize() }}
            onFocusCapture={(event) => {
                const row = (event.target as HTMLElement).closest<HTMLElement>('[data-feed-index]');
                if (row) setFocusedIndex(Number(row.dataset.feedIndex));
            }}
            onBlurCapture={(event) => {
                if (!event.currentTarget.contains(event.relatedTarget)) setFocusedIndex(null);
            }}
        >
            {virtualizer.getVirtualItems().map((item) => (
                <div
                    key={item.key}
                    ref={virtualizer.measureElement}
                    data-index={item.index}
                    data-feed-index={item.index}
                    role='listitem'
                    aria-posinset={item.index + 1}
                    aria-setsize={clips.length}
                    className='absolute left-0 top-0 w-full pb-6'
                    style={{ transform: `translateY(${item.start - scrollMargin}px)` }}
                >
                    {children(clips[item.index], item.index)}
                </div>
            ))}
        </div>
    );
}
