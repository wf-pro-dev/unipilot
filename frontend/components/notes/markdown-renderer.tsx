import React from 'react'
import { cn } from '@/lib/utils'

interface MarkdownRendererProps {
  content: string
  className?: string
  showStyles?: boolean
}

export function MarkdownRenderer({ 
  content, 
  className,
  showStyles = true 
}: MarkdownRendererProps) {
  if (!content) {
    return <div className={cn("text-muted-foreground", className)}>No content available</div>
  }

  // Clean the content to prevent XSS and rendering issues
  const cleanContent = content.trim()

  // If the content is already HTML (contains HTML tags), render it directly
  if (cleanContent.includes('<html>') || cleanContent.includes('<body>') || cleanContent.includes('<div>') || cleanContent.includes('<p>')) {
    return (
      <div 
        className={cn("note-markdown-content", className)}
        dangerouslySetInnerHTML={{ __html: cleanContent }}
      />
    )
  }

  // If it's plain markdown, display as pre-formatted text with basic styling
  return (
    <div className={cn("note-markdown-content", className)}>
      <pre className="whitespace-pre-wrap text-sm bg-gray-800/50 rounded p-4 overflow-auto text-gray-200">
        {cleanContent}
      </pre>
    </div>
  )
}

// Styled version with default markdown styles
export function StyledMarkdownRenderer({ 
  content, 
  className 
}: Omit<MarkdownRendererProps, 'showStyles'>) {
  return (
    <div className={cn("note-markdown-styled", className)} style={{ isolation: 'isolate' }}>
      <MarkdownRenderer content={content} showStyles={true} />
    </div>
  )
}

// Compact version for inline display
export function InlineMarkdownRenderer({ 
  content, 
  className 
}: Omit<MarkdownRendererProps, 'showStyles'>) {
  return (
    <div className={cn("note-markdown-inline", className)} style={{ isolation: 'isolate' }}>
      <MarkdownRenderer content={content} showStyles={false} />
    </div>
  )
} 