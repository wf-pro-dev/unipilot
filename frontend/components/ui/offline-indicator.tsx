"use client"

import { useState, useEffect } from 'react'
import { Wifi, WifiOff, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useNetworkStatus } from '@/hooks/use-network-status'
import { cn } from '@/lib/utils'

interface OfflineIndicatorProps {
  className?: string
  showText?: boolean
  variant?: 'icon' | 'badge' | 'full'
}

export function OfflineIndicator({ 
  className, 
  showText = false, 
  variant = 'icon' 
}: OfflineIndicatorProps) {
  const { isOnline, isChecking, checkNetworkStatus } = useNetworkStatus()
  const [showTooltip, setShowTooltip] = useState(false)

  // Auto-hide tooltip after 3 seconds
  useEffect(() => {
    if (showTooltip) {
      const timer = setTimeout(() => setShowTooltip(false), 3000)
      return () => clearTimeout(timer)
    }
  }, [showTooltip])

  const handleRefresh = async () => {
    setShowTooltip(true)
    await checkNetworkStatus()
  }

  if (isOnline && variant === 'icon') {
    return null // Don't show anything when online for icon variant
  }

  const getStatusText = () => {
    if (isChecking) return 'Checking connection...'
    return isOnline ? 'Online' : 'Offline - Working locally'
  }

  const getStatusIcon = () => {
    if (isChecking) {
      return <RefreshCw className="w-4 h-4 animate-spin" />
    }
    return isOnline ? <Wifi className="w-4 h-4" /> : <WifiOff className="w-4 h-4" />
  }

  const getStatusColor = () => {
    if (isChecking) return 'text-yellow-500'
    return isOnline ? 'text-green-500' : 'text-red-500'
  }

  const getBackgroundColor = () => {
    if (isChecking) return 'bg-yellow-500/10'
    return isOnline ? 'bg-green-500/10' : 'bg-red-500/10'
  }

  if (variant === 'icon') {
    return (
      <TooltipProvider>
        <Tooltip open={showTooltip} onOpenChange={setShowTooltip}>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRefresh}
              className={cn(
                'relative p-2 h-auto',
                getBackgroundColor(),
                getStatusColor(),
                'hover:bg-opacity-20 transition-all duration-200',
                className
              )}
            >
              {getStatusIcon()}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{getStatusText()}</p>
            <p className="text-xs text-gray-400">Click to refresh</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  if (variant === 'badge') {
    return (
      <div className={cn(
        'inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium',
        getBackgroundColor(),
        getStatusColor(),
        'border border-current/20',
        className
      )}>
        {getStatusIcon()}
        <span>{getStatusText()}</span>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleRefresh}
          className="p-1 h-auto ml-1 hover:bg-current/10"
        >
          <RefreshCw className={cn("w-3 h-3", isChecking && "animate-spin")} />
        </Button>
      </div>
    )
  }

  // Full variant
  return (
    <div className={cn(
      'flex items-center gap-3 p-3 rounded-lg border',
      getBackgroundColor(),
      'border-current/20',
      className
    )}>
      <div className={cn("flex items-center gap-2", getStatusColor())}>
        {getStatusIcon()}
        <span className="font-medium">{getStatusText()}</span>
      </div>
      
      {!isOnline && (
        <div className="text-sm text-gray-400">
          Your changes will be synced when you're back online
        </div>
      )}
      
      <Button
        variant="outline"
        size="sm"
        onClick={handleRefresh}
        disabled={isChecking}
        className="ml-auto"
      >
        {isChecking ? (
          <>
            <RefreshCw className="w-4 h-4 animate-spin mr-2" />
            Checking...
          </>
        ) : (
          <>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </>
        )}
      </Button>
    </div>
  )
}
