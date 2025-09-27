import { useState, useEffect, useCallback } from 'react'

interface NetworkStatus {
  online: boolean
  timestamp: number
  lastChecked: number
}

export function useNetworkStatus() {

  if (typeof window === 'undefined' || !navigator) {
    return {
      isOnline: false,
      lastChecked: 0,
      isChecking: false,
      checkNetworkStatus: () => Promise.resolve()
    }
  }

  const [networkStatus, setNetworkStatus] = useState<NetworkStatus>({
    online: navigator.onLine,
    timestamp: Date.now(),
    lastChecked: Date.now()
  })
  const [isChecking, setIsChecking] = useState(false)

  const checkNetworkStatus = useCallback(async () => {
    if (isChecking) return

    setIsChecking(true)
    try {
      // Only check browser's online status
      const browserOnline = navigator.onLine
      
      setNetworkStatus({
        online: browserOnline,
        timestamp: Date.now(),
        lastChecked: Date.now()
      })
    } catch (error) {
      console.error('Failed to check network status:', error)
      // Fall back to browser status
      setNetworkStatus(prev => ({
        ...prev,
        online: navigator.onLine,
        lastChecked: Date.now()
      }))
    } finally {
      setIsChecking(false)
    }
  }, [isChecking])

  

  useEffect(() => {
    // Initial check
    checkNetworkStatus()

    // Set up periodic checks every 5 seconds for better responsiveness
    const interval = setInterval(checkNetworkStatus, 5000)

    // Listen to browser online/offline events
    const handleOnline = async () => {
      console.log("[Network Status] Browser came online")
      setNetworkStatus(prev => ({
        ...prev,
        online: true,
        lastChecked: Date.now()
      }))
      
    }

    const handleOffline = () => {
      console.log("[Network Status] Browser offline")
      setNetworkStatus(prev => ({
        ...prev,
        online: false,
        lastChecked: Date.now()
      }))
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      clearInterval(interval)
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [checkNetworkStatus])

  return {
    isOnline: networkStatus.online,
    lastChecked: networkStatus.lastChecked,
    isChecking,
    checkNetworkStatus,
  }
}
