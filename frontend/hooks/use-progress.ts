import { useQuery, useQueryClient } from "@tanstack/react-query";

const progressKey = ['progress'] as const
// Hook to get the current upload state
export function useProgress() {
    const queryClient = useQueryClient();

    const { data: activeProgress = new Set() } = useQuery({
        queryKey: progressKey,
        queryFn: () => new Set(),
        staleTime: Infinity,
        gcTime: Infinity,
    });

    const addProgress = (progressId: string) => {
        queryClient.setQueryData(progressKey, (old: Set<string> = new Set()) => {
            const newSet = new Set(old);
            newSet.add(progressId);
            return newSet;
        });
    };

    const removeProgress = (progressId: string) => {
        queryClient.setQueryData(progressKey, (old: Set<string> = new Set()) => {
            const newSet = new Set(old);
            newSet.delete(progressId);
            return newSet;
        });
    };

    const isProgress = (progressId: string) => {
        return activeProgress.has(progressId);
    };

    return {
        activeProgress,
        addProgress,
        removeProgress,
        isProgress,
        progressCount: activeProgress.size,
    };
}