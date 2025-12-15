import React, { useEffect, useRef, useState } from 'react'
import ReactMarkdown, { Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkMath from 'remark-math'
import rehypeKatex from 'rehype-katex'
import rehypeHighlight from 'rehype-highlight'
import mermaid from 'mermaid'
import { cn } from '@/lib/utils'
import { 
  Copy, 
  Check, 
  Terminal, 
  Info, 
  AlertTriangle, 
  AlertCircle, 
  CheckCircle2, 
  Zap, 
  FileText
} from 'lucide-react'

// Import CSS for KaTeX and code highlighting
import 'katex/dist/katex.min.css'
import 'highlight.js/styles/github-dark.css' // Prefer dark theme for code blocks usually

interface MarkdownRendererProps {
  content: string
  className?: string
  variant?: 'default' | 'compact' | 'styled' | 'transparent'
}

// --- Mermaid Component ---
function MermaidDiagram({ children }: { children: string }) {
  const ref = useRef<HTMLDivElement>(null)
  const [isRendered, setIsRendered] = useState(false)

  useEffect(() => {
    if (ref.current && children && !isRendered) {
      mermaid.initialize({
        startOnLoad: false,
        theme: 'dark',
        securityLevel: 'loose',
        fontFamily: 'inherit',
      })

      const renderDiagram = async () => {
        try {
          const id = `mermaid-${Math.random().toString(36).substr(2, 9)}`
          const { svg } = await mermaid.render(id, children)
          if (ref.current) {
            ref.current.innerHTML = svg
            setIsRendered(true)
          }
        } catch (error) {
          console.error('Mermaid rendering error:', error)
          if (ref.current) {
            ref.current.innerHTML = `<div class="text-red-500 bg-red-500/10 p-4 rounded border border-red-500/20 text-sm font-mono">${children}</div>`
          }
        }
      }

      renderDiagram()
    }
  }, [children, isRendered])

  return (
    <div className="my-6 flex justify-center bg-white/5 p-4 rounded-lg overflow-x-auto">
      <div ref={ref} className="mermaid-diagram" />
    </div>
  )
}

// --- Code Block Component ---
function CodeBlock({ className, children, inline, ...props }: any) {
  const [isCopied, setIsCopied] = useState(false)
  const content = String(children).replace(/\n$/, '')
  const match = /language-(\w+)/.exec(className || '')
  const language = match ? match[1] : 'text'

  // Handle mermaid diagrams
  if (language === 'mermaid') {
    return <MermaidDiagram>{content}</MermaidDiagram>
  }

  // Check for inline code or "fake" blocks (single line, no language specified)
  // This prevents short snippets like `catch` from breaking the flow when rendered as blocks
  const isInline = inline || (!match && !content.includes('\n') && content.length < 80)

  if (isInline) {
    return (
      <code 
        className={cn(
          "bg-muted/30 px-1.5 py-0.5 rounded-md text-[0.9em] font-mono text-foreground font-medium", 
          className
        )} 
        {...props}
      >
        {children}
      </code>
    )
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(content)
    setIsCopied(true)
    setTimeout(() => setIsCopied(false), 2000)
  }

  return (
    <div className="my-6 rounded-lg overflow-hidden border border-border bg-[#0d1117] group shadow-sm">
      <div className="flex items-center justify-between px-4 py-2 bg-[#161b22] border-b border-border/10">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          {language === 'bash' || language === 'sh' || language === 'zsh' ? (
            <Terminal className="w-3.5 h-3.5" />
          ) : (
            <FileText className="w-3.5 h-3.5" />
          )}
          <span className="font-mono">{language}</span>
        </div>
        <button
          onClick={handleCopy}
          className="text-muted-foreground hover:text-foreground transition-colors p-1 rounded-md hover:bg-white/10"
          title="Copy code"
        >
          {isCopied ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
        </button>
      </div>
      <div className="relative">
        <pre className="p-4 overflow-x-auto text-sm font-mono leading-relaxed scrollbar-thin scrollbar-thumb-white/10 scrollbar-track-transparent">
          <code className={cn(className, "bg-transparent p-0 border-none")} {...props}>
            {children}
          </code>
        </pre>
      </div>
    </div>
  )
}

// --- Alert/Callout Component ---
function Blockquote({ children, ...props }: React.ComponentPropsWithoutRef<'blockquote'>) {
  // ReactMarkdown passes children as an array. We need to check the first child's text content.
  // This is a bit hacky because children structure can vary.
  // We'll do a best effort detection of the alert syntax: > [!NOTE]
  
  // Convert children to string to check for alerts (simplification)
  // In a real scenario, we'd need to inspect the ReactElement tree more carefully,
  // but often the first child is a <p> containing the text.
  
  let type: 'note' | 'tip' | 'important' | 'warning' | 'caution' | 'quote' = 'quote'
  let content = children

  // Helper to extract text from React children to check for alert type
  // This is complex with ReactNode, so we might handle it by checking if the first paragraph starts with the marker.
  // For simplicity in this renderer without deeper tree parsing, we'll style standard blockquotes nicely
  // and rely on the user to use them for emphasis. 
  
  // However, we can try to detect the pattern if it's a simple string or standard structure.
  // NOTE: implementing robust alert detection in react-markdown components is tricky without a plugin like remark-github-blockquote-alert.
  // We will style the blockquote generally to look good, and add support for "Quote" style.

  return (
    <blockquote className="my-6 pl-6 border-l-4 border-primary/30 italic text-muted-foreground bg-muted/10 py-2 rounded-r-lg" {...props}>
      {children}
    </blockquote>
  )
}

export function MarkdownRenderer({
  content,
  className,
  variant = 'default'
}: MarkdownRendererProps) {
  if (!content) {
    return <div className={cn("text-muted-foreground italic", className)}>No content available</div>
  }

  const components: Components = {
    h1: ({ children, id }) => (
      <h1 id={id} className={cn("font-bold tracking-tight text-foreground border-b border-border pb-2 first:mt-0", 
        variant === 'compact' ? "text-xl mt-4 mb-2" : "text-3xl mt-10 mb-6"
      )}>
        {children}
      </h1>
    ),
    h2: ({ children, id }) => (
      <h2 id={id} className={cn("font-semibold tracking-tight text-foreground pb-1",
        variant === 'compact' ? "text-lg mt-4 mb-2" : "text-2xl mt-10 mb-4"
      )}>
        {children}
      </h2>
    ),
    h3: ({ children, id }) => (
      <h3 id={id} className={cn("font-semibold tracking-tight text-foreground",
        variant === 'compact' ? "text-base mt-3 mb-2" : "text-xl mt-8 mb-3"
      )}>
        {children}
      </h3>
    ),
    h4: ({ children, id }) => (
      <h4 id={id} className={cn("font-medium text-foreground",
        variant === 'compact' ? "text-sm mt-2 mb-1" : "text-lg mt-6 mb-2"
      )}>
        {children}
      </h4>
    ),
    p: ({ children }) => (
      <p className={cn("text-foreground/90",
        variant === 'compact' ? "leading-relaxed [&:not(:first-child)]:mt-2" : "leading-7 [&:not(:first-child)]:mt-4"
      )}>
        {children}
      </p>
    ),
    a: ({ href, children }) => {
      const isExternal = href?.startsWith('http')
      return (
        <a 
          href={href} 
          target={isExternal ? "_blank" : undefined}
          rel={isExternal ? "noopener noreferrer" : undefined}
          className="font-medium text-primary hover:underline underline-offset-4 transition-colors"
        >
          {children}
        </a>
      )
    },
    ul: ({ children }) => (
      <ul className="my-6 ml-6 list-disc [&>li]:mt-2 marker:text-muted-foreground/60">
        {children}
      </ul>
    ),
    ol: ({ children }) => (
      <ol className="my-6 ml-6 list-decimal [&>li]:mt-2 marker:text-muted-foreground/60">
        {children}
      </ol>
    ),
    li: ({ children }) => (
      <li className="pl-2">{children}</li>
    ),
    blockquote: Blockquote,
    code: CodeBlock,
    pre: ({ children }) => <>{children}</>, // Handled by code component
    hr: () => <hr className="my-8 border-border" />,
    table: ({ children }) => (
      <div className="my-6 w-full overflow-y-auto rounded-lg border border-border">
        <table className="w-full text-sm">
          {children}
        </table>
      </div>
    ),
    thead: ({ children }) => (
      <thead className="bg-muted/50 border-b border-border text-left">
        {children}
      </thead>
    ),
    th: ({ children }) => (
      <th className="px-4 py-3 font-medium text-muted-foreground">
        {children}
      </th>
    ),
    tr: ({ children }) => (
      <tr className="border-b border-border last:border-0 even:bg-muted/20 hover:bg-muted/40 transition-colors">
        {children}
      </tr>
    ),
    td: ({ children }) => (
      <td className="px-4 py-3 align-top">
        {children}
      </td>
    ),
    img: ({ src, alt }) => (
      <div className="my-8">
        <div className="relative rounded-lg overflow-hidden border border-border shadow-sm bg-muted/20">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img 
            src={src} 
            alt={alt} 
            className="w-full h-auto object-cover"
            loading="lazy"
          />
        </div>
        {alt && (
          <p className="mt-2 text-sm text-center text-muted-foreground italic">
            {alt}
          </p>
        )}
      </div>
    ),
    // Math components are handled by remark-math/rehype-katex but wrapped in spans/divs
  }

  return (
    <div className={cn("w-full max-w-none", className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={[
          rehypeKatex,
          rehypeHighlight
        ]}
        components={components}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}

export function StyledMarkdownRenderer(props: Omit<MarkdownRendererProps, 'variant'>) {
  return <MarkdownRenderer {...props} variant="styled" />
}

export function InlineMarkdownRenderer(props: Omit<MarkdownRendererProps, 'variant'>) {
  return (
    <MarkdownRenderer 
      {...props} 
      variant="compact" 
      className={cn("text-sm", props.className)}
    />
  )
}
