import { FlatList, FlatListProps } from "../core/flatlist";
import { models } from "@/wailsjs/go/models";
import { useCallback, useMemo, useState } from "react";
import { UserItem } from "./user-item";
import { GlassCard } from "../ui/glass-card";
import { EmptyState as EmptyStateComponent } from "../ui/empty-state";
import { Users } from "lucide-react";
import { useClusterUsersScroll, useUsersScroll } from "@/hooks/use-users";
import { SearchList } from "../core/searchlist";

export type ClusterUsersListProps = Partial<FlatListProps<models.User>> & Pick<FlatListProps<models.User>, 'userID'> & {
    courseID: string;
} ;


export function ClusterUsersList({
    userID,
    courseID,
    containerClassName,
    renderItem: renderItemProp,
    numColumns = 3,
    itemsPerPage = 20,
}: ClusterUsersListProps) {


    const renderItem = useCallback((user: models.User) => {
        if (user.ID === userID) return null;
        return <UserItem user={user} />
    }, []);

    const EmptyState = useMemo(() => {
        return (
            <div className="flex flex-1">
                <GlassCard
                    variant="board"
                    className="flex-1 items-center"
                >
                    <EmptyStateComponent
                        icon={Users}
                        title="No Users found"
                        description="No Users found"
                        className="flex-1 items-center"
                    />
                </GlassCard>
            </div >
        )
    }, [userID]);


    return (
        <SearchList
            key={"cluster-users-" + courseID}
            entityID={courseID}
            itemsPerPage={itemsPerPage}
            useScroll={useClusterUsersScroll}
            renderItem={renderItemProp || renderItem}
            keyExtractor={(user: models.User) => user.ID}
            numColumns={numColumns}
            containerClassName={containerClassName}
            emptyState={EmptyState}

            searchConfig={{
                placeholder: "Search users...",
                searchableFields: ["Username", "Email"]
            }}
        />
    )

}