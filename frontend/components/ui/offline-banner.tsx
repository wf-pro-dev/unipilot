"use client"

import { WifiOff, RefreshCw, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useNetworkStatus } from '@/hooks/use-network-status'
import { cn } from '@/lib/utils'

interface OfflineBannerProps {
  className?: string
  variant?: 'warning' | 'info' | 'error'
  showSyncInfo?: boolean
}

export function OfflineBanner({ 
  className, 
  variant = 'warning',
  showSyncInfo = true 
}: OfflineBannerProps) {
  const { isOnline, isChecking, checkNetworkStatus } = useNetworkStatus()

  if (isOnline) {
    return null
  }

  const getVariantStyles = () => {
    switch (variant) {
      case 'error':
        return {
          background: 'bg-red-500/10 border-red-500/20',
          text: 'text-red-200',
          icon: 'text-red-400',
          button: 'bg-red-500/20 hover:bg-red-500/30 text-red-200'
        }
      case 'info':
        return {
          background: 'bg-blue-500/10 border-blue-500/20',
          text: 'text-blue-200',
          icon: 'text-blue-400',
          button: 'bg-blue-500/20 hover:bg-blue-500/30 text-blue-200'
        }
      default: // warning
        return {
          background: 'bg-yellow-500/10 border-yellow-500/20',
          text: 'text-yellow-200',
          icon: 'text-yellow-400',
          button: 'bg-yellow-500/20 hover:bg-yellow-500/30 text-yellow-200'
        }
    }
  }

  const styles = getVariantStyles()

  return (
    <div className={cn(
      'flex items-center justify-between p-4 border-b',
      styles.background,
      'border-current/20',
      className
    )}>
      <div className="flex items-center gap-3">
        <div className={cn("flex items-center gap-2", styles.icon)}>
          {isChecking ? (
            <RefreshCw className="w-5 h-5 animate-spin" />
          ) : (
            <WifiOff className="w-5 h-5" />
          )}
        </div>
        
        <div className={cn("flex flex-col", styles.text)}>
          <div className="flex items-center gap-2">
            <span className="font-medium">
              {isChecking ? 'Checking connection...' : 'You are currently offline'}
            </span>
            {!isChecking && (
              <AlertTriangle className="w-4 h-4" />
            )}
          </div>
          
          {showSyncInfo && !isChecking && (
            <p className="text-sm opacity-80">
              Your changes will be automatically synced when you're back online
            </p>
          )}
        </div>
      </div>

      <Button
        variant="outline"
        size="sm"
        onClick={checkNetworkStatus}
        disabled={isChecking}
        className={cn(
          'border-current/30',
          styles.button
        )}
      >
        {isChecking ? (
          <>
            <RefreshCw className="w-4 h-4 animate-spin mr-2" />
            Checking...
          </>
        ) : (
          <>
            <RefreshCw className="w-4 h-4 mr-2" />
            Try Again
          </>
        )}
      </Button>
    </div>
  )
}
