import React, { useMemo } from 'react';
import { motion, useScroll, useTransform } from 'framer-motion';
import { cn } from '@/lib/utils';

interface ScrollProps<T> {
    data: T[]
    renderItem: (item: T) => React.JSX.Element
    keyExtractor?: (item: T, index: number) => string | number;
    numColumns: number
    containerClassName?: string
}

export function Scroll<T>({
    data,
    renderItem,
    keyExtractor,
    numColumns,
    containerClassName
}: ScrollProps<T>) {
    const scrollContainerRef = React.useRef<HTMLDivElement>(null);

    const rows = useMemo(() => {
        const result = [];
        for (let i = 0; i < data.length; i += numColumns) {
            result.push(data.slice(i, i + numColumns));
        }
        return result;
    }, [data, numColumns])

    return (
        <div 
            ref={scrollContainerRef} 
            className="flex flex-col overflow-y-auto snap-y snap-mandatory space-y-4 min-h-0 flex-1"
        >
            {rows.map((row, index) => (
                <Row<T> 
                    key={index} 
                    rowData={row} 
                    renderItem={renderItem} 
                    containerClassName={containerClassName} 
                    scrollContainerRef={scrollContainerRef}
                />
            ))}
        </div>
    );
}

interface RowProps<T> {
    rowData: T[]
    renderItem: (item: T) => React.JSX.Element
    containerClassName?: string
    scrollContainerRef: React.RefObject<HTMLDivElement>
}

const Row = <T,>({
    rowData,
    renderItem,
    containerClassName,
    scrollContainerRef
}: RowProps<T>) => {
    return (
        <motion.div
            initial={{ opacity: 0.5 }}
            whileInView={{ opacity: 1 }}
            viewport={{ 
                root: scrollContainerRef, // Use scroll container as viewport root!
                amount: 0.8 
            }}
            transition={{ duration: 0.5, ease: "easeOut" }}
            className={cn("snap-start flex-shrink-0 flex items-center justify-center", containerClassName)}
        >
            {rowData.map((item, index) => (
                <React.Fragment key={index}>
                    {renderItem(item)}
                </React.Fragment>
            ))}
        </motion.div>
    );
};