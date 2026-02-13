/**
 * Shared TypeScript types for search and filter components
 * 
 * These types are used by both SearchFilter (client-side) and 
 * SearchFilterServer (server-side) to ensure consistency.
 */

/**
 * Represents a single option in a filter dropdown
 */
export interface FilterOption {
    label: string  // Display text shown to user
    value: string  // Value used internally and in callbacks
}

/**
 * Defines a filter configuration
 * 
 * @template T - The data type being filtered
 */
export interface FilterDefinition<T = any> {
    field: keyof T           // The field in the data object to filter on
    label: string            // Display label for the filter
    type: 'select'           // Filter type (currently only 'select', extensible for future types)
    placeholder?: string     // Placeholder text for the dropdown
    width?: string           // Tailwind width class (e.g., 'w-36', 'w-48')

    /**
     * Extract unique filter options from the data array
     * Use this when you want to auto-generate options from your data
     */
    extractOptions?: (data: T[]) => string[]

    /**
     * Provide custom filter options
     * Use this when you have predefined options (especially for server-side filtering)
     */
    customOptions?: FilterOption[]
}

/**
 * Configuration for the search input
 * 
 * @template T - The data type being searched
 */
export interface SearchConfig<T = any> {
    placeholder: string          // Placeholder text for search input
    searchableFields: (keyof T)[] // Fields to search across (supports nested objects)
    enabled?: boolean            // Whether to show the search input (default: true)
}

/**
 * Current state of all filters
 * 
 * Maps filter field names to their selected values
 * 
 * @example
 * {
 *   Status: "active",
 *   Priority: "high",
 *   Role: "all"
 * }
 */
export interface FilterState {
    [key: string]: string
}