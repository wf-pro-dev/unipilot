"use client"

import { FieldValues, UseFormReturn } from "react-hook-form"
import { FormMessage } from "@/components/ui/form"
import { shouldShowError, ErrorDisplayConfig, FieldState } from "./use-form-error-display"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

interface FormErrorMessageProps<TFieldValues extends FieldValues> {
    fieldState: FieldState<TFieldValues>
    formState: UseFormReturn<TFieldValues>["formState"]
    config: ErrorDisplayConfig
    className?: string
    onErrorResolved?: () => void
}

/**
 * Reusable FormMessage component with configurable error display strategy
 * 
 * @example
 * ```tsx
 * <FormErrorMessage
 *     fieldState={fieldState}
 *     formState={form.formState}
 *     config={{ strategy: "onTouched" }}
 *     className="text-xs text-red-400 ml-1"
 * />
 * ```
 */
export function FormErrorMessage<TFieldValues extends FieldValues>({
    fieldState,
    formState,
    config,
    className,
    onErrorResolved,
}: FormErrorMessageProps<TFieldValues>) {
    const showError = shouldShowError(fieldState, formState, config)

    if (!showError || !fieldState.error) {
        return null
    }
 
    toast.error(fieldState.error.message)
    onErrorResolved?.()
    return null
}
