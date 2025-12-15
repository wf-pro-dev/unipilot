/**
 * Loading component for the courses page.
 * 
 * Returns null because the courses page handles its own loading state
 * internally with a loading spinner. This component exists to satisfy
 * Next.js App Router's loading.tsx file requirement but doesn't render
 * any skeleton UI.
 * 
 * @returns {null} Always returns null
 */
export default function Loading() {
  return null
}
