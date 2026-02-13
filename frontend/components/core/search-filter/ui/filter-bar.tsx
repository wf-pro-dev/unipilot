"use client"

import { X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { GlassCard } from "@/components/ui/glass-card"
import { CardContent } from "@/components/ui/card"
import { SearchInput } from "./search-input"
import { FilterDropdown } from "./filter-dropdown"
import { SearchConfig, FilterDefinition, FilterState, FilterOption } from "../types"    

export interface FilterBarProps<T = any> {
  // Search
  searchTerm: string
  onSearchChange: (value: string) => void
  searchConfig?: SearchConfig<T>
  
  // Filters
  filters: FilterState
  onFilterChange: (field: string, value: string) => void
  filterDefinitions: FilterDefinition<T>[]
  filterOptions: Record<string, FilterOption[]>
  
  // Actions
  onClearAll?: () => void
  hasActiveFilters: boolean
  
  // Layout
  layout?: 'horizontal' | 'vertical'
  
  // State
  isLoading?: boolean
}

/**
 * FilterBar - Composed UI component
 * 
 * Combines SearchInput + FilterDropdowns into a cohesive layout.
 * Shared across client-side and server-side filtering.
 * Contains only presentation logic - no business logic.
 * 
 * @example
 * <FilterBar
 *   searchTerm={search}
 *   onSearchChange={setSearch}
 *   searchConfig={{ placeholder: "Search...", searchableFields: ["name"] }}
 *   filters={{ status: "active" }}
 *   onFilterChange={(field, value) => setFilters({ ...filters, [field]: value })}
 *   filterDefinitions={[
 *     { field: "status", label: "Status", type: "select" }
 *   ]}
 *   filterOptions={{ status: [{ label: "Active", value: "active" }] }}
 *   hasActiveFilters={true}
 *   onClearAll={() => clearFilters()}
 * />
 */
export function FilterBar<T = any>({
  searchTerm,
  onSearchChange,
  searchConfig,
  filters,
  onFilterChange,
  filterDefinitions,
  filterOptions,
  onClearAll,
  hasActiveFilters,
  layout = 'horizontal',
  isLoading = false,
}: FilterBarProps<T>) {
  const layoutClass = layout === 'horizontal' 
    ? "flex-col lg:flex-row lg:items-center" 
    : "flex-col"

  return (
    <GlassCard variant="board" className="flex-grow-0 flex-row">
      <CardContent className="flex-1 p-2">
        <div className={`flex ${layoutClass} space-x-2`}>
          
          {/* Search Input */}
          {searchConfig?.enabled !== false && (
            <SearchInput
              value={searchTerm}
              onChange={onSearchChange}
              placeholder={searchConfig?.placeholder}
              disabled={isLoading}
            />
          )}

          {/* Filter Dropdowns + Clear Button */}
          {filterDefinitions.length > 0 && (
            <div className="flex items-center gap-2 m-0 space-y-0">
              {filterDefinitions.map((filterDef) => {
                const fieldKey = filterDef.field as string
                const options = filterOptions[fieldKey] || []
                
                return (
                  <FilterDropdown
                    key={fieldKey}
                    value={filters[fieldKey] || "all"}
                    onChange={(value) => onFilterChange(fieldKey, value)}
                    options={options}
                    placeholder={filterDef.placeholder}
                    label={filterDef.label}
                    width={filterDef.width}
                    disabled={isLoading}
                  />
                )
              })}

              {/* Clear Filters Button */}
              {hasActiveFilters && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onClearAll}
                  className="h-10 px-3 text-gray-400 hover:text-white hover:bg-white/10"
                  disabled={isLoading}
                >
                  <X className="h-4 w-4" />
                </Button>
              )}
              
              
            </div>
          )}
        </div>
      </CardContent>
    </GlassCard>
  )
}