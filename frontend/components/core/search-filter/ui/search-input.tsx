"use client"

import { Search } from "lucide-react"
import { Input } from "@/components/ui/input"

export interface SearchInputProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
  className?: string
}

/**
 * SearchInput - Pure UI component for search input
 * 
 * Shared across client-side and server-side filtering.
 * Contains no business logic - just UI and callbacks.
 * 
 * @example
 * <SearchInput
 *   value={searchTerm}
 *   onChange={(value) => setSearchTerm(value)}
 *   placeholder="Search items..."
 * />
 */
export function SearchInput({
  value,
  onChange,
  placeholder = "Search...",
  disabled = false,
  className = ""
}: SearchInputProps) {
  return (
    <div className={`relative flex-1 ${className}`}>
      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
      <Input
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="pl-10 bg-white/5 border-white/10 rounded-xl transition-all duration-300 h-10"
        disabled={disabled}
        autoCapitalize="off"
      />
    </div>
  )
}