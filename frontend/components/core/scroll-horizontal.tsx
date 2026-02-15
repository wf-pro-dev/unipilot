import React, { useMemo, useCallback, useRef, useEffect } from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';
import { PageResponse } from '@/types/models';

export interface ScrollHorizontalProps<T> {
    data: PageResponse<T>
    renderItem: (item: T) => React.JSX.Element | null
    keyExtractor?: (item: T, index: number) => string | number;
    numColumns: number // Number of items per row (horizontally)
    numRows?: number // Number of rows visible (vertically stacked)
    containerClassName?: string // for the outer scroll container
    rowClassName?: string // for each individual row

    // Infinite scroll props
    onLoadMore?: () => void
    isFetchingMore?: boolean
    prefetchDistance?: number // Distance from right edge (in pixels) to trigger load

    // Empty state
    emptyState?: React.JSX.Element;
}

export function ScrollHorizontal<T>({
    data,
    renderItem,
    keyExtractor,
    numColumns,
    numRows,
    containerClassName,
    rowClassName,
    onLoadMore,
    isFetchingMore = false,
    prefetchDistance = 800,
    emptyState
}: ScrollHorizontalProps<T>) {
    const scrollContainerRef = useRef<HTMLDivElement>(null);
    const loadingRef = useRef(false);

    const list = data.Data

    // Chunk data into rows (same as vertical scroll.tsx)
    const rows = useMemo(() => {
        const result = [];
        for (let i = 0; i < list.length; i += numColumns) {
            result.push(list.slice(i, i + numColumns));
        }
        return result;
    }, [list, numColumns])

    // Group rows into pages if numRows is specified
    const pages = useMemo(() => {
        if (!numRows) return rows.map(row => [row]); // Each row is its own page
        
        const result = [];
        for (let i = 0; i < rows.length; i += numRows) {
            result.push(rows.slice(i, i + numRows));
        }
        return result;
    }, [rows, numRows]);

    // Infinite scroll handler (horizontal)
    const handleScroll = useCallback(() => {
        const container = scrollContainerRef.current;
        if (!container || !onLoadMore || !data.HasMore || isFetchingMore || loadingRef.current) {
            return;
        }

        const { scrollLeft, scrollWidth, clientWidth } = container;
        const distanceFromRight = scrollWidth - (scrollLeft + clientWidth);

        // Trigger load when within prefetchDistance of right edge
        if (distanceFromRight < prefetchDistance) {
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
            className={cn("flex flex-row overflow-x-auto snap-x snap-mandatory space-x-4 min-w-0 flex-1",
                containerClassName
            )}
        >
            {pages.map((pageRows, pageIndex) => (
                <Page<T>
                    key={pageIndex}
                    pageRows={pageRows}
                    numColumns={numColumns}
                    renderItem={renderItem}
                    rowClassName={rowClassName}
                    scrollContainerRef={scrollContainerRef}
                    keyExtractor={keyExtractor}
                />
            ))}

            {/* Loading indicator */}
            {isFetchingMore && (
                <div className="flex items-center justify-center px-4">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                </div>
            )}
        </div>
    );
}

interface PageProps<T> {
    pageRows: T[][]
    numColumns: number
    renderItem: (item: T) => React.JSX.Element | null
    rowClassName?: string
    scrollContainerRef: React.RefObject<HTMLDivElement>
    keyExtractor?: (item: T, index: number) => string | number;
}

const Page = <T,>({
    pageRows,
    numColumns,
    renderItem,
    rowClassName,
    scrollContainerRef,
    keyExtractor
}: PageProps<T>) => {
    return (
        <motion.div
            initial={{ opacity: 0.5 }}
            whileInView={{ opacity: 1 }}
            viewport={{
                root: scrollContainerRef,
                amount: 0.8
            }}
            transition={{ duration: 0.5, ease: "easeOut" }}
            className="snap-start flex-shrink-0 flex flex-col space-y-4 w-full"
        >
            {pageRows.map((row, rowIndex) => (
                <div
                    key={rowIndex}
                    className={cn("grid items-center justify-center overflow-visible",
                        `grid-cols-${numColumns}`,
                        rowClassName
                    )}
                >
                    {row.map((item, index) => (
                        <React.Fragment key={index}>
                            {renderItem(item)}
                        </React.Fragment>
                    ))}
                </div>
            ))}
        </motion.div>
    );
};