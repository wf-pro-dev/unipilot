"use client"

import { useMemo, useState, useEffect, useCallback } from "react"
import { useDebounce } from "use-debounce"
import { FilterBar } from "./ui/filter-bar"
import { SearchConfig, FilterDefinition, FilterState, FilterOption } from "./types"
import { PageResponse } from "@/types/models"

export interface SearchFilterServerProps<T> {
  // Data from server (paginated)
  data: PageResponse<T>
  
  // Search configuration
  searchConfig?: SearchConfig<T>
  
  // Filter definitions
  filterDefinitions?: FilterDefinition<T>[]
  
  // Initial values (from URL params)
  initialFilters?: FilterState
  initialSearchTerm?: string
  
  // ✅ CLEAN SEPARATION: Separate callbacks for search and filters
  onSearchChange?: (debouncedSearch: string) => void  // Debounced search only
  onFilterChange?: (filters: FilterState) => void      // Filters only (no debounce needed)
  onClearAll?: () => void
  
  // Server state
  isLoading?: boolean
  isFetching?: boolean
  
  // Styling
  className?: string
  layout?: 'horizontal' | 'vertical'
  
  // Debounce config
  debounceMs?: number
  
  // Children - render function
  children: (data: PageResponse<T>) => React.ReactNode
}

/**
 * SearchFilterServer - Server-side filtering component
 * 
 * Clean separation of concerns:
 * - Search: Debounced, triggers onSearchChange after 500ms idle
 * - Filters: Immediate, triggers onFilterChange instantly
 * 
 * @example
 * ```tsx
 * function UserDirectory() {
 *   const [search, setSearch] = useState("")
 *   const [filters, setFilters] = useState({})
 *   
 *   const { data, isLoading } = useUsersInfinite({
 *     searchTerm: search,
 *     filters: filters
 *   })
 *   
 *   return (
 *     <SearchFilterServer
 *       data={data}
 *       searchConfig={{ placeholder: "Search...", searchableFields: [...] }}
 *       filterDefinitions={[...]}
 *       onSearchChange={(debouncedSearch) => setSearch(debouncedSearch)}
 *       onFilterChange={(filters) => setFilters(filters)}
 *       isLoading={isLoading}
 *     >
 *       {(data) => <UserList data={data} />}
 *     </SearchFilterServer>
 *   )
 * }
 * ```
 */
export function SearchFilterServer<T>({
  data,
  searchConfig,
  filterDefinitions = [],
  initialFilters = {},
  initialSearchTerm = "",
  onSearchChange,
  onFilterChange,
  onClearAll,
  isLoading = false,
  isFetching = false,
  className = "",
  layout = "horizontal",
  debounceMs = 500,
  children
}: SearchFilterServerProps<T>) {

  const [searchTerm, setSearchTerm] = useState(initialSearchTerm)
  const [debouncedSearchTerm] = useDebounce(searchTerm, debounceMs)


  // ✅ Trigger search callback when debounced value changes
  useEffect(() => {
    onSearchChange?.(debouncedSearchTerm)
  }, [debouncedSearchTerm, onSearchChange])


  const [filters, setFilters] = useState<FilterState>(initialFilters)


  // ✅ Trigger filter callback when filters change (no debounce)
  useEffect(() => {
    onFilterChange?.(filters)
  }, [filters, onFilterChange])


  const filterOptions = useMemo(() => {
    const options: Record<string, FilterOption[]> = {}
    
    filterDefinitions.forEach((filterDef) => {
      if (filterDef.customOptions) {
        options[filterDef.field as string] = filterDef.customOptions
      } else if (filterDef.extractOptions) {
        const values = filterDef.extractOptions(data.Data)
        options[filterDef.field as string] = values.map(v => ({ label: v, value: v }))
      } else {
        const uniqueValues = Array.from(
          new Set(data.Data.map((item: T) => item[filterDef.field]).filter(Boolean))
        )
        options[filterDef.field as string] = uniqueValues.map(v => ({ 
          label: String(v), 
          value: String(v) 
        }))
      }
    })
    
    return options
  }, [data, filterDefinitions])


  
  // Search handler - just updates state, debouncing + callback handled by useEffect
  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value)
  }, [])

  // Filter handler - just updates state, callback handled by useEffect
  const handleFilterChange = useCallback((field: string, value: string) => {
    setFilters(prev => ({ ...prev, [field]: value }))
  }, [])

  // Clear all handler
  const handleClearAll = useCallback(() => {
    setSearchTerm("")
    setFilters({})
    onClearAll?.()
  }, [onClearAll])

  const hasActiveFilters = 
    searchTerm !== "" || 
    Object.values(filters).some(v => v && v !== "all")


  return (
    <div className={`flex flex-col h-full min-h-0 space-y-4 ${className}`}>
      {/* Search & Filter Bar */}
      <FilterBar
        searchTerm={searchTerm}
        onSearchChange={handleSearchChange}
        searchConfig={searchConfig}
        filters={filters}
        onFilterChange={handleFilterChange}
        filterDefinitions={filterDefinitions}
        filterOptions={filterOptions}
        onClearAll={handleClearAll}
        hasActiveFilters={hasActiveFilters}
        layout={layout}
        isLoading={isLoading}
      />

      {/* Render data */}
      <div className="flex h-full min-h-0">
        {children(data)}
      </div>
    </div>
  )
}