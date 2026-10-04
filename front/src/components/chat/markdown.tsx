import { createContext, useContext, type ComponentProps } from 'react'
import ReactMarkdown, { type Components, type ExtraProps } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { cn } from '@/lib/utils'
import { remarkLineBreaks } from './remark-line-breaks'

// Markdown is the one renderer of chat messages, in every chat: GitHub
// Flavored Markdown, links that open in a new tab, and styled code, code
// blocks and tables. keepLines keeps the lines someone typed, which Markdown
// would otherwise join.
export function Markdown({
  text,
  keepLines = false,
}: {
  text: string
  keepLines?: boolean
}) {
  return (
    <ReactMarkdown
      remarkPlugins={keepLines ? KEEP_LINES_PLUGINS : REPLY_PLUGINS}
      components={COMPONENTS}
    >
      {text}
    </ReactMarkdown>
  )
}

// MarkdownText draws a text part of a reply.
export function MarkdownText({ text }: { text: string }) {
  return <Markdown text={text} />
}

// UserMarkdownText draws a text part of the user's message.
export function UserMarkdownText({ text }: { text: string }) {
  return <Markdown text={text} keepLines />
}

const REPLY_PLUGINS = [remarkGfm]
const KEEP_LINES_PLUGINS = [remarkGfm, remarkLineBreaks]

// InCodeBlock tells the code of a fenced block from inline code: a block
// without a language has no class name to tell them apart.
const InCodeBlock = createContext(false)

const COMPONENTS: Components = {
  a: MarkdownLink,
  code: MarkdownCode,
  pre: MarkdownPre,
  table: MarkdownTable,
  th: MarkdownTableHeader,
  td: MarkdownTableCell,
  p: ({ children }) => (
    <p className='my-2 max-w-[72ch] first:mt-0 last:mb-0'>{children}</p>
  ),
  ul: ({ children }) => (
    <ul className='my-2 max-w-[72ch] list-disc space-y-1 pl-5 first:mt-0 last:mb-0'>
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className='my-2 max-w-[72ch] list-decimal space-y-1 pl-5 first:mt-0 last:mb-0'>
      {children}
    </ol>
  ),
  li: ({ children }) => <li className='pl-1'>{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className='my-2 max-w-[72ch] border-l-2 border-border pl-3 text-muted-foreground italic first:mt-0 last:mb-0'>
      {children}
    </blockquote>
  ),
  hr: () => <hr className='my-3 border-border' />,
  h1: ({ children }) => (
    <h1 className='my-3 max-w-[72ch] text-lg font-semibold first:mt-0 last:mb-0'>
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className='my-3 max-w-[72ch] text-base font-semibold first:mt-0 last:mb-0'>
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className='my-2 max-w-[72ch] text-sm font-semibold first:mt-0 last:mb-0'>
      {children}
    </h3>
  ),
}

function MarkdownLink({
  node: _node,
  className: _className,
  ...props
}: ComponentProps<'a'> & ExtraProps) {
  return (
    <a
      {...props}
      target='_blank'
      rel='noopener noreferrer'
      className='font-medium underline underline-offset-2 hover:opacity-80'
    />
  )
}

function MarkdownCode({
  node: _node,
  className,
  ...props
}: ComponentProps<'code'> & ExtraProps) {
  const inBlock = useContext(InCodeBlock)
  if (inBlock) {
    return <code {...props} className={cn('font-mono text-xs', className)} />
  }
  return (
    <code
      {...props}
      className='rounded bg-muted px-1 py-0.5 font-mono text-[0.85em] text-foreground'
    />
  )
}

function MarkdownPre({
  node: _node,
  className: _className,
  ...props
}: ComponentProps<'pre'> & ExtraProps) {
  return (
    <InCodeBlock.Provider value>
      <pre
        {...props}
        className='my-2 max-w-full scrollbar-thin overflow-x-auto rounded-md border border-border/70 bg-muted/35 p-3 text-foreground first:mt-0 last:mb-0'
      />
    </InCodeBlock.Provider>
  )
}

function MarkdownTable({
  node: _node,
  className: _className,
  ...props
}: ComponentProps<'table'> & ExtraProps) {
  return (
    <div className='my-3 max-w-full scrollbar-thin overflow-x-auto rounded-md border border-border/70 first:mt-0 last:mb-0'>
      <table
        {...props}
        className='w-full min-w-[42rem] border-separate border-spacing-0 text-left text-xs'
      />
    </div>
  )
}

function MarkdownTableHeader({
  node: _node,
  className: _className,
  ...props
}: ComponentProps<'th'> & ExtraProps) {
  return (
    <th
      {...props}
      className='border-b border-border bg-muted/55 px-3 py-2 font-semibold whitespace-nowrap text-foreground'
    />
  )
}

function MarkdownTableCell({
  node: _node,
  className: _className,
  ...props
}: ComponentProps<'td'> & ExtraProps) {
  return (
    <td
      {...props}
      className='border-b border-border/50 px-3 py-2 align-top leading-5'
    />
  )
}
