import React, { useEffect, useRef } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeKatex from 'rehype-katex'
import rehypeHighlight from 'rehype-highlight'
import mermaid from 'mermaid'
import { cn } from '@/lib/utils'

// Import CSS for KaTeX and code highlighting
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github.css'

interface MarkdownRendererProps {
  content: string
  className?: string
  variant?: 'default' | 'compact' | 'styled'
}

// Custom Mermaid component with proper spacing
function MermaidDiagram({ children }: { children: string }) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (ref.current && children) {
      mermaid.initialize({ 
        startOnLoad: true,
        theme: 'default',
        securityLevel: 'loose',
      })
      
      const renderDiagram = async () => {
        try {
          const { svg } = await mermaid.render(`mermaid-${Date.now()}`, children)
          if (ref.current) {
            ref.current.innerHTML = svg
          }
        } catch (error) {
          console.error('Mermaid rendering error:', error)
          if (ref.current) {
            ref.current.innerHTML = `<pre class="text-red-500 bg-red-50 p-4 rounded border">${children}</pre>`
          }
        }
      }
      
      renderDiagram()
    }
  }, [children])

  return (
    <div className="my-8 flex justify-center">
      <div ref={ref} className="mermaid-diagram max-w-full overflow-x-auto" />
    </div>
  )
}

export function MarkdownRenderer({ 
  content, 
  className,
  variant = 'default'
}: MarkdownRendererProps) {
  if (!content) {
    return <div className={cn("text-muted-foreground", className)}>No content available</div>
  }

  const baseClasses = "prose prose-sm max-w-none"
  const variantClasses = {
    default: "prose-gray dark:prose-invert",
    compact: "prose-gray dark:prose-invert prose-compact",
    styled: "prose-blue dark:prose-invert prose-lg"
  }

  return (
    <div className={cn(baseClasses, variantClasses[variant], className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={[
          rehypeKatex,
          rehypeHighlight
        ]}
        components={{
          // Custom components for better styling with proper spacing
          h1: ({ children }) => (
            <h1 className="text-2xl font-bold text-foreground border-b border-border pb-2 mb-6 mt-8 first:mt-0">
              {children}
            </h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-xl font-semibold text-foreground border-b border-border pb-2 mb-5 mt-7">
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-lg font-medium text-foreground mb-4 mt-6">
              {children}
            </h3>
          ),
          code: ({ inline, className, children, ...props }) => {
            const match = /language-(\w+)/.exec(className || '')
            const language = match ? match[1] : ''
            
            // Handle mermaid diagrams
            if (language === 'mermaid') {
              return <MermaidDiagram>{String(children).replace(/\n$/, '')}</MermaidDiagram>
            }
            
            if (inline) {
              return (
                <code 
                  className="bg-muted px-1.5 py-0.5 rounded text-sm font-mono text-muted-foreground"
                  {...props}
                >
                  {children}
                </code>
              )
            }
            return (
              <div className="my-6">
                <code className={cn("block", className)} {...props}>
                  {children}
                </code>
              </div>
            )
          },
          pre: ({ children, ...props }) => (
            <div className="my-8">
              <pre 
                className="bg-muted rounded-lg p-4 overflow-x-auto text-sm font-mono border"
                {...props}
              >
                {children}
              </pre>
            </div>
          ),
          blockquote: ({ children }) => (
            <blockquote className="border-l-4 border-primary pl-6 py-2 my-6 italic text-muted-foreground bg-muted/20 rounded-r">
              {children}
            </blockquote>
          ),
          table: ({ children }) => (
            <div className="my-8 overflow-x-auto rounded-lg border">
              <table className="min-w-full border-collapse">
                {children}
              </table>
            </div>
          ),
          th: ({ children }) => (
            <th className="border-b border-border bg-muted px-4 py-3 text-left font-semibold">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="border-b border-border px-4 py-3">
              {children}
            </td>
          ),
          // Add spacing around math equations
          div: ({ className, children, ...props }) => {
            // KaTeX equations have specific classes
            if (className?.includes('math-display')) {
              return (
                <div className={cn("my-8 flex justify-center", className)} {...props}>
                  {children}
                </div>
              )
            }
            return <div className={className} {...props}>{children}</div>
          },
          p: ({ children }) => (
            <p className="mb-4 leading-7">
              {children}
            </p>
          )
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}

// Styled version with enhanced visual hierarchy
export function StyledMarkdownRenderer({ 
  content, 
  className 
}: Omit<MarkdownRendererProps, 'variant'>) {
  return (
    <MarkdownRenderer 
      content={content} 
      className={className}
      variant="styled"
    />
  )
}

// Compact version for inline display
export function InlineMarkdownRenderer({ 
  content, 
  className 
}: Omit<MarkdownRendererProps, 'variant'>) {
  return (
    <MarkdownRenderer 
      content={content} 
      className={className}
      variant="compact"
    />
  )
} 