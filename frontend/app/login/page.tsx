"use client"

import AuthPage from "@/components/auth/page"
import { useRouter } from "next/navigation"

/**
 * Login page component for user authentication.
 * 
 * Provides a dedicated login route that renders the login form and handles
 * post-authentication navigation. After successful login, redirects users to
 * the dashboard (home page).
 * 
 * Features:
 * - Renders LoginForm component with success callback
 * - Client-side navigation after successful authentication
 * - Redirects to dashboard on login success
 * 
 * @returns {JSX.Element} The login page with LoginForm component
 */
export default function LoginPage() {
  const router = useRouter()

  /**
   * Handles successful login by redirecting to the dashboard.
   * 
   * Called by LoginForm component after successful authentication.
   * Uses client-side navigation to redirect without full page reload.
   */
  const handleLoginSuccess = () => {
    // Redirect to dashboard after successful login
    router.push("/")
  }

  return (
    <AuthPage onLoginSuccess={handleLoginSuccess} />
  )
} 
