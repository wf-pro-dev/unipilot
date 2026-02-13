import React, { useMemo, useCallback, useRef, useEffect } from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { PageResponse } from '@/types/models';

export interface ScrollProps<T> {
    data: PageResponse<T>
    renderItem: (item: T) => React.JSX.Element | null
    keyExtractor?: (item: T, index: number) => string | number;
    numColumns: number
    containerClassName?: string

    // Infinite scroll props
    onLoadMore?: () => void
    isFetchingMore?: boolean
    prefetchDistance?: number // Distance from bottom (in pixels) to trigger load

    // Empty state
    emptyState?: React.JSX.Element;
}

export function Scroll<T>({
    data,
    renderItem,
    keyExtractor,
    numColumns,
    containerClassName,
    onLoadMore,
    isFetchingMore = false,
    prefetchDistance = 800,
    emptyState
}: ScrollProps<T>) {
    const scrollContainerRef = useRef<HTMLDivElement>(null);
    const loadingRef = useRef(false);

    const list = data.Data;

    const rows = useMemo(() => {
        const result = [];
        for (let i = 0; i < list.length; i += numColumns) {
            result.push(list.slice(i, i + numColumns));
        }
        return result;
    }, [list, numColumns])

    // Infinite scroll handler
    const handleScroll = useCallback(() => {
        const container = scrollContainerRef.current;
        if (!container || !onLoadMore || !data.HasMore || isFetchingMore || loadingRef.current) {
            return;
        }

        const { scrollTop, scrollHeight, clientHeight } = container;
        const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);

        // Trigger load when within prefetchDistance of bottom
        if (distanceFromBottom < prefetchDistance) {
            loadingRef.current = true;
            onLoadMore();
        }
    }, [onLoadMore, data.HasMore, isFetchingMore, prefetchDistance]);

    // Reset loading ref when fetch completes
    useEffect(() => {
        if (!isFetchingMore) {
            loadingRef.current = false;
        }
    }, [isFetchingMore]);

    // Attach scroll listener
    useEffect(() => {
        const container = scrollContainerRef.current;
        if (!container) return;

        container.addEventListener('scroll', handleScroll, { passive: true });

        // Also check on mount in case content doesn't fill viewport
        handleScroll();

        return () => {
            container.removeEventListener('scroll', handleScroll);
        };
    }, [handleScroll]);

    if (list.length === 0) {
        return emptyState;
    }

    return (
        <div
            ref={scrollContainerRef}
            className="flex flex-col overflow-y-auto snap-y snap-mandatory space-y-4 min-h-0 flex-1"
        >

            {rows.map((row, index) => (
                <Row<T>
                    key={index}
                    rowData={row}
                    numColumns={numColumns}
                    renderItem={renderItem}
                    containerClassName={containerClassName}
                    scrollContainerRef={scrollContainerRef}
                    keyExtractor={keyExtractor}
                />
            ))}

            {/* Loading indicator */}
            {isFetchingMore && (
                <div className="flex items-center justify-center py-4">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                </div>
            )}

           
        </div>
    );
}

interface RowProps<T> {
    rowData: T[]
    numColumns: number
    renderItem: (item: T) => React.JSX.Element | null
    containerClassName?: string
    scrollContainerRef: React.RefObject<HTMLDivElement>
    keyExtractor?: (item: T, index: number) => string | number;
}

const Row = <T,>({
    rowData,
    numColumns,
    renderItem,
    containerClassName,
    scrollContainerRef,
    keyExtractor
}: RowProps<T>) => {
    return (
        <motion.div
            initial={{ opacity: 0.5 }}
            whileInView={{ opacity: 1 }}
            viewport={{
                root: scrollContainerRef,
                amount: 0.8
            }}
            transition={{ duration: 0.5, ease: "easeOut" }}
            className={cn("snap-start flex-shrink-0 items-center justify-center overflow-visible",
                `grid grid-cols-${numColumns}`,
                containerClassName)}
        >
            {rowData.map((item, index) => (
                <React.Fragment key={index}>
                    {renderItem(item)}
                </React.Fragment>
            ))}
        </motion.div>
    );
};