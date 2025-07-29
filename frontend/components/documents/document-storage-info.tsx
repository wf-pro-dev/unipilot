"use client"

import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Progress } from "@/components/ui/progress"
import { Badge } from "@/components/ui/badge"
import { 
  HardDrive, 
  Info,
  RefreshCw
} from "lucide-react"
import { useUserStorageInfo } from "@/hooks/use-documents"

export function DocumentStorageInfo() {
  const { data: storageInfo, isLoading, refetch } = useUserStorageInfo()

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 B"
    const k = 1024
    const sizes = ["B", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i]
  }

  const getStoragePercentage = (): number => {
    if (!storageInfo) return 0
    const maxStorage = 2 * 1024 * 1024 * 1024 // 2GB in bytes
    return Math.min((storageInfo.TotalSize / maxStorage) * 100, 100)
  }

  const getStorageColor = (): string => {
    const percentage = getStoragePercentage()
    if (percentage >= 90) return "text-red-500"
    if (percentage >= 75) return "text-yellow-500"
    return "text-blue-500"
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" size="sm" className="h-8">
          <HardDrive className="h-4 w-4 mr-1" />
          Storage
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <div className="space-y-4">
          {/* Header */}
          <div className="flex items-center justify-between">
            <h4 className="font-semibold">Storage Usage</h4>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => refetch()}
              disabled={isLoading}
              className="h-8 w-8 p-0"
            >
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>

          {storageInfo ? (
            <>
              {/* Storage Bar */}
              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span>Used Storage</span>
                  <span className="font-medium">
                    {formatFileSize(storageInfo.TotalSize)} / 2 GB
                  </span>
                </div>
                <Progress value={getStoragePercentage()} className="h-2" />
                <div className={`text-xs text-center ${getStorageColor()}`}>
                  {Math.round(getStoragePercentage())}% used
                </div>
              </div>

              {/* Statistics */}
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-primary">
                    {storageInfo.DocumentCount}
                  </div>
                  <div className="text-xs text-muted-foreground">Documents</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-primary">
                    {formatFileSize(storageInfo.TotalSize)}
                  </div>
                  <div className="text-xs text-muted-foreground">Used</div>
                </div>
              </div>

              {/* Storage Levels */}
              <div className="space-y-2">
                {getStoragePercentage() >= 90 && (
                  <Badge variant="destructive" className="w-full justify-center">
                    Storage Almost Full
                  </Badge>
                )}
                {getStoragePercentage() >= 75 && getStoragePercentage() < 90 && (
                  <Badge variant="secondary" className="w-full justify-center bg-yellow-500/20 text-yellow-700">
                    Storage Getting Full
                  </Badge>
                )}
              </div>

              {/* Last Updated */}
              <div className="text-xs text-muted-foreground text-center">
                Last updated: {new Date(storageInfo.LastCalculatedAt).toLocaleString()}
              </div>
            </>
          ) : isLoading ? (
            <div className="flex items-center justify-center py-4">
              <RefreshCw className="h-4 w-4 animate-spin mr-2" />
              <span className="text-sm">Loading storage info...</span>
            </div>
          ) : (
            <div className="flex items-center justify-center py-4">
              <Button variant="outline" onClick={() => refetch()} size="sm">
                <Info className="h-4 w-4 mr-2" />
                Load Storage Info
              </Button>
            </div>
          )}

          {/* Storage Limits Info */}
          <div className="border-t pt-3">
            <div className="text-xs text-muted-foreground space-y-1">
              <div className="font-medium">Storage Limits:</div>
              <div>• Max file size: 50 MB</div>
              <div>• Max per assignment: 200 MB</div>
              <div>• Total limit: 2 GB</div>
            </div>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
} 