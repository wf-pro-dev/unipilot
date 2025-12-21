"use client"

import { LoginForm } from "@/components/auth/login-form"
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
    <div className="">
      {/* Floating background elements */}
      <div className="absolute top-20 left-10 w-72 h-72 bg-blue-500/10 rounded-full blur-3xl animate-float"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl animate-float-delayed"></div>
      
      <div className="relative z-10">
        <LoginForm onLoginSuccess={handleLoginSuccess} />
      </div>
    </div>
  )
} 