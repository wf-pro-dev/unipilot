import { FlatList, FlatListProps } from "../core/flatlist";
import { models } from "@/wailsjs/go/models";
import { useCallback, useMemo } from "react";
import { UserItem } from "./user-item";
import { GlassCard } from "../ui/glass-card";
import { EmptyState as EmptyStateComponent } from "../ui/empty-state";
import { Users } from "lucide-react";
import { useUsersScroll } from "@/hooks/use-users";
import { SearchList } from "../core/searchlist";

export type UserListProps = Partial<FlatListProps<models.User>> & Pick<FlatListProps<models.User>, 'userID'>;

export function UserList({
    userID,
    containerClassName,
}: UserListProps) {

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
            key={"users-" + userID}
            userID={userID}
            itemsPerPage={20}
            useScroll={useUsersScroll}
            renderItem={renderItem}
            keyExtractor={(user: models.User) => user.ID}
            numColumns={3}
            containerClassName={containerClassName}
            emptyState={EmptyState}

            searchConfig={{
                placeholder: "Search users...",
                searchableFields: ["Username", "Email"]
            }}
            filterDefinitions={[
                {
                    field: "University",
                    label: "University",
                    type: "select",
                    extractOptions: (data: models.User[]) => {
                        return Array.from(new Set(data.map(user => user.University)))
                    }
                }
            ]}
        />
    )

}