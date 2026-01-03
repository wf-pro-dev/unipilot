import { UseFormReturn, FieldValues } from "react-hook-form"

/**
 * Field state type matching react-hook-form's getFieldState return type
 */
export interface FieldState<TFieldValues extends FieldValues> {
    error?: { message?: string }
    isTouched: boolean
    isDirty: boolean
    invalid: boolean
}

/**
 * Error display strategies for controlling when validation errors are shown
 */
export type ErrorDisplayStrategy =
    | "always"           // Always show errors when they exist
    | "onTouched"        // Show errors only after field is touched
    | "onBlur"           // Show errors only after field loses focus
    | "onSubmit"         // Show errors only after form submission attempt
    | "onStepAttempt"    // Show errors only after step validation attempt (for multi-step forms)
    | "onDirty"          // Show errors only after field has been modified

/**
 * Configuration for error display behavior
 */
export interface ErrorDisplayConfig {
    /** The strategy to use for displaying errors */
    strategy: ErrorDisplayStrategy
    /** Current step number (for multi-step forms) */
    step?: number
    /** Whether current step validation was attempted */
    stepAttempted?: boolean
    /** Whether form submission was attempted */
    submitAttempted?: boolean
}

/**
 * Determines whether to show a field error based on the configured strategy
 * 
 * @param fieldState - The field state from react-hook-form
 * @param formState - The form state from react-hook-form
 * @param config - Configuration for error display strategy
 * @returns boolean indicating whether to show the error
 * 
 * @example
 * ```ts
 * const showError = shouldShowError(fieldState, form.formState, {
 *   strategy: "onTouched"
 * })
 * ```
 */
export function shouldShowError<TFieldValues extends FieldValues>(
    fieldState: FieldState<TFieldValues>,
    formState: UseFormReturn<TFieldValues>["formState"],
    config: ErrorDisplayConfig
): boolean {
    const { error } = fieldState
    if (!error) return false

    const { strategy, stepAttempted, submitAttempted } = config

    switch (strategy) {
        case "always":
            return true

        case "onTouched":
            return fieldState.isTouched

        case "onBlur":
            return fieldState.isTouched && !fieldState.isDirty

        case "onSubmit":
            return submitAttempted ?? formState.submitCount > 0

        case "onStepAttempt":
            return stepAttempted  ?? false

        case "onDirty":
            return fieldState.isDirty

        default:
            return false
    }
}

/**
 * Hook for managing error display in multi-step forms
 * Provides centralized control over when errors are displayed
 * 
 * @param form - The form instance from react-hook-form
 * @param currentStep - Current step number (optional)
 * @param stepAttempted - Whether current step validation was attempted (optional)
 * @returns Configuration object for error display
 * 
 * @example
 * ```tsx
 * const { getErrorDisplayChecker } = useFormErrorDisplay(form, step, step1Attempted)
 * 
 * const shouldShow = getErrorDisplayChecker("onStepAttempt")
 * ```
 */
export function useFormErrorDisplay<TFieldValues extends FieldValues>(
    form: UseFormReturn<TFieldValues>,
    currentStep?: number,
    stepAttempted?: boolean
) {
    const formState = form.formState
    const submitAttempted = formState.submitCount > 0

    /**
     * Creates a function to check if error should be shown for a specific field
     * 
     * @param strategy - The error display strategy to use
     * @param customStepAttempted - Override for stepAttempted (optional)
     * @returns Function that takes fieldState and returns boolean
     */
    const getErrorDisplayChecker = (
        strategy: ErrorDisplayStrategy,
        customStepAttempted?: boolean
    ) => {
        return (fieldState: FieldState<TFieldValues>) => {
            return shouldShowError(fieldState, formState, {
                strategy,
                step: currentStep,
                stepAttempted: customStepAttempted ?? stepAttempted,
                submitAttempted,
            })
        }
    }

    return {
        getErrorDisplayChecker,
        submitAttempted,
        stepAttempted: stepAttempted ?? false,
        formState,
    }
}

