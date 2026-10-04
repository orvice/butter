import { describe, expect, it } from 'vitest'
import { breakLines, type MarkdownNode } from './remark-line-breaks'

const paragraph = (...children: MarkdownNode[]): MarkdownNode => ({
  type: 'paragraph',
  children,
})
const text = (value: string): MarkdownNode => ({ type: 'text', value })
const br: MarkdownNode = { type: 'break' }

function broken(tree: MarkdownNode): MarkdownNode {
  breakLines(tree)
  return tree
}

describe('breakLines', () => {
  it('turns each newline in text into a line break', () => {
    expect(
      broken({ type: 'root', children: [paragraph(text('one\ntwo\r\nthree'))] })
    ).toEqual({
      type: 'root',
      children: [paragraph(text('one'), br, text('two'), br, text('three'))],
    })
  })

  it('reaches text inside emphasis, links and lists', () => {
    const tree = {
      type: 'root',
      children: [
        {
          type: 'list',
          children: [
            {
              type: 'listItem',
              children: [
                paragraph({ type: 'strong', children: [text('bold\nline')] }),
              ],
            },
          ],
        },
      ],
    }
    expect(broken(tree).children?.[0].children?.[0].children?.[0]).toEqual(
      paragraph({
        type: 'strong',
        children: [text('bold'), br, text('line')],
      })
    )
  })

  it('leaves code alone, and text without newlines as it was', () => {
    const tree = {
      type: 'root',
      children: [
        { type: 'code', value: 'a\nb' },
        paragraph(text('same'), { type: 'inlineCode', value: 'x\ny' }),
      ],
    }
    expect(broken(structuredClone(tree))).toEqual(tree)
  })

  it('keeps empty lines as consecutive breaks', () => {
    expect(broken(paragraph(text('a\n\nb')))).toEqual(
      paragraph(text('a'), br, br, text('b'))
    )
  })
})
