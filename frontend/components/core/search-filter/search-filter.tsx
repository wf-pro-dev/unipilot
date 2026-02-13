"use client"

import { useMemo, useState, useEffect } from "react"
import { FilterBar } from "./ui/filter-bar"
import { SearchConfig, FilterDefinition, FilterState, FilterOption } from "./types"

export interface SearchFilterProps<T> {
  // Data
  data: T[]
  
  // Search configuration
  searchConfig?: SearchConfig<T>
  
  // Filter definitions
  filterDefinitions?: FilterDefinition<T>[]
  
  // Initial values (from URL params or other source)
  initialFilters?: FilterState
  initialSearchTerm?: string
  
  // Callbacks - parent handles routing/side effects
  onFilterChange?: (filters: FilterState) => void
  onSearchChange: (searchTerm: string) => void
  onClearAll?: () => void
  
  // Styling
  className?: string
  layout?: 'horizontal' | 'vertical'
  
  // Optional
  isLoading?: boolean
  
  // Children - render function for filtered data
  children: (filteredData: T[]) => React.ReactNode
}

/**
 * SearchFilter - Client-side filtering component
 * 
 * Filters data in-memory using useMemo for performance.
 * All filtering happens instantly on the client side.
 * 
 * **Use this component when:**
 * - Data is fully loaded (< 10,000 items)
 * - Instant filtering is desired
 * - No server round-trips needed
 * - Examples: Assignments, Courses, small datasets
 * 
 * **Features:**
 * - ✅ Instant results (no debouncing)
 * - ✅ Filters in-memory with useMemo
 * - ✅ No loading states needed
 * - ✅ Simple state management
 * 
 * Uses shared FilterBar UI component for consistent design.
 * 
 * @example
 * ```tsx
 * <SearchFilter
 *   data={assignments}
 *   searchConfig={{
 *     placeholder: "Search assignments...",
 *     searchableFields: ["Title", "Course"]
 *   }}
 *   filterDefinitions={[
 *     { field: "Status", label: "Status", type: "select" }
 *   ]}
 *   onFilterChange={(filters) => updateURL(filters)}
 *   onSearchChange={(search) => console.log(search)}
 * >
 *   {(filteredData) => <DataGrid data={filteredData} />}
 * </SearchFilter>
 * ```
 */
export function SearchFilter<T extends Record<string, any>>({
  data,
  searchConfig,
  filterDefinitions = [],
  initialFilters = {},
  initialSearchTerm = "",
  onFilterChange,
  onSearchChange,
  onClearAll,
  className = "",
  layout = "horizontal",
  isLoading = false,
  children
}: SearchFilterProps<T>) {
  // Internal state - updates immediately for responsive UI
  const [searchTerm, setSearchTerm] = useState(initialSearchTerm)
  const [filters, setFilters] = useState<FilterState>(initialFilters)

  // Sync with initial values when they change (e.g., URL params change from browser back button)
  useEffect(() => {
    setSearchTerm(initialSearchTerm)
  }, [initialSearchTerm])

  useEffect(() => {
    setFilters(initialFilters)
  }, [initialFilters])

  // Extract unique options for each filter from the data
  const filterOptions = useMemo(() => {
    const options: Record<string, FilterOption[]> = {}
    
    filterDefinitions.forEach((filterDef) => {
      if (filterDef.customOptions) {
        // Use provided custom options
        options[filterDef.field as string] = filterDef.customOptions
      } else if (filterDef.extractOptions) {
        // Use custom extraction function
        const values = filterDef.extractOptions(data)
        options[filterDef.field as string] = values.map(v => ({ label: v, value: v }))
      } else {
        // Default: extract unique values from the field
        const uniqueValues = Array.from(
          new Set(data.map(item => item[filterDef.field]).filter(Boolean))
        )
        options[filterDef.field as string] = uniqueValues.map(v => ({ 
          label: String(v), 
          value: String(v) 
        }))
      }
    })
    
    return options
  }, [data, filterDefinitions])

  // Apply filters and search (CLIENT-SIDE - happens instantly)
  const filteredData = useMemo(() => {
    return data.filter((item) => {
      // Search filter - checks if item matches search term
      if (searchConfig?.enabled !== false && searchTerm) {
        const matchesSearch = searchConfig?.searchableFields.some(field => {
          const value = item[field]
          if (value === null || value === undefined) return false
          
          // Handle nested objects (e.g., Course.Name)
          if (typeof value === 'object' && value !== null) {
            return Object.values(value).some(v => 
              String(v).toLowerCase().includes(searchTerm.toLowerCase())
            )
          }
          
          return String(value).toLowerCase().includes(searchTerm.toLowerCase())
        })
        
        if (!matchesSearch) return false
      }

      // Filter dropdowns - check if item matches all active filters
      const matchesFilters = filterDefinitions.every((filterDef) => {
        const filterValue = filters[filterDef.field as string]
        if (!filterValue || filterValue === "all") return true
        
        const itemValue = item[filterDef.field]
        
        // Handle nested objects (e.g., Course.Code)
        if (typeof itemValue === 'object' && itemValue !== null) {
          return Object.values(itemValue).some(v => String(v) === filterValue)
        }
        
        return String(itemValue) === filterValue
      })

      return matchesFilters
    })
  }, [data, searchTerm, filters, searchConfig, filterDefinitions])

  // Handlers
  const handleSearchChange = (value: string) => {
    setSearchTerm(value)
    onSearchChange(value)
  }

  const handleFilterChange = (field: string, value: string) => {
    const newFilters = { ...filters, [field]: value }
    setFilters(newFilters)
    onFilterChange?.(newFilters)
  }

  const handleClearAll = () => {
    setSearchTerm("")
    setFilters({})
    onClearAll?.()
  }

  const hasActiveFilters = 
    searchTerm !== "" || 
    Object.values(filters).some(v => v && v !== "all")

  return (
    <div className={`flex flex-col h-full min-h-0 space-y-4 ${className}`}>
      {/* Search & Filter Bar - Shared UI Component */}
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

      {/* Render filtered data */}
      <div className="flex h-full min-h-0">
        {children(filteredData)}
      </div>
    </div>
  )
}