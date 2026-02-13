"use client"

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { GlassCard } from "@/components/ui/glass-card"
import { FilterOption } from "../types"

export interface FilterDropdownProps {
  value: string
  onChange: (value: string) => void
  options: FilterOption[]
  placeholder?: string
  label?: string
  width?: string
  disabled?: boolean
}

/**
 * FilterDropdown - Pure UI component for filter dropdown
 * 
 * Shared across client-side and server-side filtering.
 * Contains no business logic - just UI and callbacks.
 * 
 * @example
 * <FilterDropdown
 *   value={selectedStatus}
 *   onChange={(value) => setStatus(value)}
 *   options={[
 *     { label: "Active", value: "active" },
 *     { label: "Inactive", value: "inactive" }
 *   ]}
 *   label="Status"
 *   placeholder="All Statuses"
 * />
 */
export function FilterDropdown({
  value,
  onChange,
  options,
  placeholder,
  label,
  width = "w-36",
  disabled = false
}: FilterDropdownProps) {
  return (
    <Select value={value} onValueChange={onChange} disabled={disabled}>
      <SelectTrigger className={`${width} bg-white/5 border-white/10 h-10 rounded-xl`}>
        <SelectValue placeholder={placeholder || label} />
      </SelectTrigger>
      <SelectContent className="bg-transparent border-none">
        <GlassCard variant="board">
          <SelectItem 
            value="all" 
            className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
          >
            {placeholder || `All ${label}`}
          </SelectItem>
          {options.map((option) => (
            <SelectItem
              key={option.value}
              value={option.value}
              className="text-gray-300 focus:text-white focus:bg-white/10 cursor-pointer"
            >
              {option.label}
            </SelectItem>
          ))}
        </GlassCard>
      </SelectContent>
    </Select>
  )
}