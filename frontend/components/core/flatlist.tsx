import React, { useMemo } from 'react';
import { Scroll } from '@/components/core/scroll';
import { useFriendsScroll } from '@/hooks/use-friends';
import { models } from '@/wailsjs/go/models';
import { PageResponse } from '@/types/models';
import { UserItem } from '../community/user-item';

import { Users } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { GlassCard } from '../ui/glass-card';
import { UseInfiniteQueryResult, InfiniteData } from '@tanstack/react-query';
import { cn } from '@/lib/utils';
import { ScrollProps } from './scroll';

export interface FlatListProps<T> extends Omit<ScrollProps<T>, 'data'> {
    userID: string;
    itemsPerPage: number;
    useScroll: (props: {limit?: number, userID: string}) => UseInfiniteQueryResult<InfiniteData<PageResponse<T>, unknown>, Error>;
    emptyState: React.JSX.Element;

}

export function FlatList<T>({
    userID,
    itemsPerPage,
    useScroll,
    renderItem,
    keyExtractor,
    numColumns,
    containerClassName,
    emptyState,
}: FlatListProps<T>) {

    if (!userID) { return null; }

    const {
        data,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        isLoading,
        isError,
        error
    } = useScroll({limit: itemsPerPage, userID: userID});

    // Flatten all pages into a single PageResponse for the Scroll component
    const flattenedData: PageResponse<T> = useMemo(() => {
        if (!data) {
            return { Data: [], HasMore: false, Cursor: undefined };
        }
        // Combine all pages into single array
        const allPages = data.pages.flatMap(page => page.Data || []);;
        // Use the last page's metadata
        const lastPage = data.pages[data.pages.length - 1];

        return {
            Data: allPages,
            HasMore: lastPage.HasMore,
            Cursor: lastPage.Cursor
        };
    }, [data]);

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
            </div>
        );
    }

    if (isError) {
        return (
            <div className="flex items-center justify-center h-full">
                <p className="text-destructive">
                    Error loading friends: {error instanceof Error ? error.message : 'Unknown error'}
                </p>
            </div>
        );
    }

    if (flattenedData.Data.length === 0) {
        return emptyState;
    }

    return (
        
        <Scroll<T>
            data={flattenedData}
            renderItem={renderItem}
            keyExtractor={keyExtractor}
            numColumns={numColumns}
            containerClassName={containerClassName}
            onLoadMore={() => {
                if (hasNextPage && !isFetchingNextPage) {
                    fetchNextPage();
                }
            }}
            isFetchingMore={isFetchingNextPage}
            prefetchDistance={800} // Load more when 800px from bottom
        />
    );
}
