import React, { useEffect, useState, useRef } from 'react'
import { cn } from '@/lib/utils'
import { ChevronRight } from 'lucide-react'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'

interface TocItem {
  id: string
  text: string
  level: number
  children?: TocItem[]
}

interface MarkdownTableOfContentsProps {
  markdown: string
  containerRef: React.RefObject<HTMLElement>
  className?: string
  maxLevel?: number // e.g., 3 means show h1, h2, h3 only
  defaultCollapsed?: boolean
}

export function MarkdownTableOfContents({
  markdown,
  containerRef,
  className,
  maxLevel = 4,
  defaultCollapsed = true,
}: MarkdownTableOfContentsProps) {
  const [tocItems, setTocItems] = useState<TocItem[]>([])
  const [activeId, setActiveId] = useState<string>('')
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set())
  const observerRef = useRef<IntersectionObserver | null>(null)

  // Extract headings from markdown and build hierarchical structure
  useEffect(() => {
    const flatHeadings: Array<{ id: string; text: string; level: number }> = []
    const lines = markdown.split('\n')

    lines.forEach((line) => {
      const match = line.match(/^(#{1,6})\s+(.+)$/)
      if (match) {
        const level = match[1].length
        if (level <= maxLevel) {
          const text = match[2].trim()
          // Generate ID the same way react-markdown does
          const id = text.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
          flatHeadings.push({ id, text, level })
        }
      }
    })

    // Build hierarchical tree
    const buildTree = (items: typeof flatHeadings): TocItem[] => {
      const tree: TocItem[] = []
      const stack: TocItem[] = []

      items.forEach((item) => {
        const tocItem: TocItem = { ...item, children: [] }

        // Find parent in stack
        while (stack.length > 0 && stack[stack.length - 1].level >= item.level) {
          stack.pop()
        }

        if (stack.length === 0) {
          // Top level item
          tree.push(tocItem)
        } else {
          // Child item
          const parent = stack[stack.length - 1]
          if (!parent.children) parent.children = []
          parent.children.push(tocItem)
        }

        stack.push(tocItem)
      })

      return tree
    }

    const main_toc = flatHeadings[0]
    const tree = buildTree(flatHeadings.slice(1, flatHeadings.length)) // Remove first item (h1)
    setTocItems([main_toc, ...tree])

    // Initialize collapsed state for all parent items
    if (defaultCollapsed) {
      const getAllParentIds = (items: TocItem[]): string[] => {
        const ids: string[] = []
        items.forEach((item) => {
          if (item.children && item.children.length > 0) {
            ids.push(item.id)
            ids.push(...getAllParentIds(item.children))
          }
        })
        return ids
      }
      setCollapsedSections(new Set(getAllParentIds(tree)))
    }
  }, [markdown, maxLevel, defaultCollapsed])

  // Track scroll position and update active heading
  useEffect(() => {
    if (!containerRef.current || tocItems.length === 0) return

    // Cleanup previous observer
    if (observerRef.current) {
      observerRef.current.disconnect()
    }

    // Flatten tree to get all IDs
    const flattenTree = (items: TocItem[]): string[] => {
      const ids: string[] = []
      items.forEach((item) => {
        ids.push(item.id)
        if (item.children) {
          ids.push(...flattenTree(item.children))
        }
      })
      return ids
    }

    const allIds = flattenTree(tocItems)

    const observer = new IntersectionObserver(
      (entries) => {
        // Find the first intersecting entry
        const intersecting = entries.find(entry => entry.isIntersecting)
        if (intersecting) {
          setActiveId(intersecting.target.id)
        }
      },
      {
        root: containerRef.current,
        rootMargin: '-20% 0px -70% 0px',
        threshold: [0, 0.25, 0.5, 0.75, 1],
      }
    )

    observerRef.current = observer

    // Observe all headings with a small delay to ensure they're rendered
    const timeoutId = setTimeout(() => {
      allIds.forEach((id) => {
        // Use getElementById or attribute selector to avoid CSS selector issues with IDs starting with numbers
        const element = document.getElementById(id) ||
          containerRef.current?.querySelector(`[id="${id}"]`)
        if (element) {
          observer.observe(element)
        }
      })
    }, 100)

    return () => {
      clearTimeout(timeoutId)
      observer.disconnect()
    }
  }, [tocItems, containerRef])

  const scrollToSection = (id: string) => {
    if (!containerRef.current) return

    // Use getElementById or escape the selector properly
    const element = document.getElementById(id) ||
      containerRef.current.querySelector(`[id="${id}"]`)
    if (element) {
      const container = containerRef.current
      const elementTop = (element as HTMLElement).offsetTop
      const offset = 100 // Offset from top

      container.scrollTo({
        top: elementTop - offset,
        behavior: 'smooth',
      })
    }
  }

  const toggleSection = (id: string) => {
    setCollapsedSections((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  if (tocItems.length === 0) return null

  // Recursive component to render tree items
  const renderTocItems = (items: TocItem[], parentLevel = 0) => {
    return items.map((item) => {
      const isActive = activeId === item.id
      const hasChildren = item.children && item.children.length > 0
      const isCollapsed = collapsedSections.has(item.id)

      if (hasChildren) {
        // Parent item with children - use Collapsible
        return (
          <li key={item.id} className="space-y-1">
            <Collapsible
              open={!isCollapsed}
              onOpenChange={() => toggleSection(item.id)}
            >
              <div
                className={cn(
                  "flex items-center gap-1 w-full text-left text-sm  px-2  rounded-md transition-all duration-200",
                  "hover:bg-white/5 hover:text-foreground",
                  "border-l-2 -ml-px",
                  isActive
                    ? "text-white  bg-white/10 "
                    : "text-muted-foreground border-transparent hover:border-white/10"
                )} >

                {/* Click on text to scroll */}
                <button
                  onClick={() => scrollToSection(item.id)}
                  className="flex-1 py-1 text-left"
                  title="Jump to section"
                >

                  <span className="text-caption line-clamp-1 leading-snug flex-1">
                    {item.text}
                  </span>

                </button>

                <CollapsibleTrigger asChild>
                  <div role="button" className="text-muted-foreground hover:text-white hover:bg-white/10 rounded-sm cursor-pointer group-data-[collapsible=icon]:hidden">
                    <ChevronRight 
                    width={16} 
                    height={16} 
                    className={cn("ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90", isCollapsed ? "rotate-0" : "rotate-90")} />
                  </div>
                </CollapsibleTrigger>

              </div>

              <CollapsibleContent className="pl-4 border-l border-white/5 ml-2 mt-1" >
                <ul className="space-y-1">
                  {renderTocItems(item.children || [])}
                </ul>
              </CollapsibleContent>

            </Collapsible>
          </li>
        )
      } else {
        // Leaf item without children
        return (
          <li key={item.id}>
            <button
              onClick={() => scrollToSection(item.id)}
              className={cn(
                "text-left text-sm w-full py-1 px-2 rounded-md transition-all duration-200",
                "hover:bg-white/5 hover:text-foreground",
                isActive
                  ? "text-white  bg-white/10 "
                  : "text-muted-foreground border-transparent hover:border-white/10"
              )}
              title={item.text}
            >
              <span className="text-caption line-clamp-2 leading-snug">
                {item.text}
              </span>
            </button>
          </li>
        )
      }
    })
  }

  return (
    <div className={cn("space-y-2", className)}>
      <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3 px-3">
        On This Page
      </h3>

      <nav className="group">
        <ul className="space-y-1">
          {renderTocItems(tocItems)}
        </ul>
      </nav>
    </div>
  )
}