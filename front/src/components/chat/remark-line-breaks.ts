// The parts of a Markdown syntax tree (mdast) the plugin touches.
export interface MarkdownNode {
  type: string
  value?: string
  children?: MarkdownNode[]
}

const NEWLINE = /\r\n|\r|\n/

// remarkLineBreaks keeps the lines someone typed: every newline inside text
// becomes a line break, where Markdown would join the lines into one. Code
// keeps its own newlines, since inline code and code blocks are not text.
export function remarkLineBreaks() {
  return (tree: MarkdownNode) => {
    breakLines(tree)
  }
}

// breakLines splits each text node under node at its newlines, in place.
export function breakLines(node: MarkdownNode): void {
  if (!node.children) return
  const children: MarkdownNode[] = []
  for (const child of node.children) {
    if (child.type !== 'text' || !child.value || !NEWLINE.test(child.value)) {
      breakLines(child)
      children.push(child)
      continue
    }
    child.value.split(NEWLINE).forEach((line, i) => {
      if (i > 0) children.push({ type: 'break' })
      if (line) children.push({ type: 'text', value: line })
    })
  }
  node.children = children
}
