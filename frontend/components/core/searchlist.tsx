import React, { useCallback, useMemo, useState } from 'react';
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
import { SearchFilterServer, SearchFilterServerProps } from './search-filter/search-filter-server';
import { FilterState } from './search-filter/types';

export interface SearchListProps<T> extends Omit<ScrollProps<T>, 'data'>, Omit<SearchFilterServerProps<T>, 'data' | 'children' | 'onFilterChange'> {
    
    userID: string;
    itemsPerPage: number;
    useScroll: (props: { limit?: number, userID: string, search?: string, filters?: FilterState }) => UseInfiniteQueryResult<InfiniteData<PageResponse<T>, unknown>, Error>;

    emptyState?: React.JSX.Element;

}

export function SearchList<T>({
    userID,
    itemsPerPage,
    useScroll,
    renderItem,
    keyExtractor,
    numColumns,
    containerClassName,
    emptyState,
    
    searchConfig,
    filterDefinitions,


}: SearchListProps<T>) {

    if (!userID) { return null; }
    const [filters, setFilters] = useState<FilterState>({});
    const [searchTerm, setSearchTerm] = useState("");

    const {
        data,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
        isLoading,
        isError,
        error
    } = useScroll({ limit: itemsPerPage, userID: userID, search: searchTerm, filters: filters });

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

    const onFilterChange = useCallback((filters: FilterState) => {
        setFilters(filters)
    }, []);

    const onSearchChange = useCallback((searchTerm: string) => {
        setSearchTerm(searchTerm)
    }, []);


    return (
        <SearchFilterServer<T>
            data={flattenedData}
            searchConfig={searchConfig}
            filterDefinitions={filterDefinitions}
            onFilterChange={onFilterChange}
            onSearchChange={onSearchChange}
            isLoading={isLoading}
            debounceMs={500}
        >
            {(data) => (
                <Scroll<T>
                    data={data}
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
                    emptyState={emptyState}
                />
            )}
        </SearchFilterServer>

    );
}
