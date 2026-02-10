import { FlatList, FlatListProps } from "../core/flatlist";
import { models } from "@/wailsjs/go/models";
import { useFriendsScroll } from "@/hooks/use-friends";
import { useCallback, useMemo } from "react";
import { UserItem } from "./user-item";
import { GlassCard } from "../ui/glass-card";
import { EmptyState as EmptyStateComponent } from "../ui/empty-state";
import { Users } from "lucide-react";
import { useRouter } from "next/navigation";

export type FriendListProps = Partial<FlatListProps<models.User>> & Pick<FlatListProps<models.User>, 'userID'>;

export function FriendList({
    userID,
    containerClassName,
    itemsPerPage = 20,
    numColumns = 3,
    renderItem: CustomRenderItem,
}: FriendListProps) {

    const router = useRouter();
    const renderItem = useCallback((user: models.User) => {
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
                        title="No friends found"
                        description="No friends found"
                        className="flex-1 items-center"
                        buttonText="Add Friend"
                        onClick={() => {
                            router.push(`/community/friends/${userID}`);
                        }}
                    />
                </GlassCard>
            </div >
        )
    }, [userID]);


    return (
        <FlatList
            key={"friends-" + userID}
            userID={userID}
            itemsPerPage={itemsPerPage}
            useScroll={useFriendsScroll}
            renderItem={CustomRenderItem || renderItem}
            keyExtractor={(user: models.User) => user.ID}
            numColumns={numColumns}
            containerClassName={containerClassName}
            emptyState={EmptyState}
        />
    )

}