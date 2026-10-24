import {
  Button,
  Card,
  ChoicePicker,
  Column,
  Divider,
  Row,
  Text,
  TextField,
} from '@a2ui/react/v0_9'
import { describe, expect, it } from 'vitest'
import { butterCatalog } from './catalog'
import { CATALOG_ID } from './protocol'

describe('butter-basic-v1 catalog', () => {
  it('reuses the official layouts without changing the catalog identity', () => {
    expect(butterCatalog.id).toBe(CATALOG_ID)
    for (const component of [Card, Column, Divider]) {
      expect(butterCatalog.components.get(component.name)).toBe(component)
    }
    // Row only adds responsive presentation around the official view/render;
    // its API and schemas stay those of the SDK, including weight bindings.
    expect(butterCatalog.components.get('Row')?.schema).toBe(Row.schema)
  })

  it('keeps Butter text and server-owned form controls', () => {
    for (const official of [Text, TextField, ChoicePicker, Button]) {
      const implementation = butterCatalog.components.get(official.name)
      expect(implementation).toBeDefined()
      expect(implementation).not.toBe(official)
    }
    // These props connect server field errors to the local form draft; they
    // are Butter additions, not part of the official input APIs.
    expect(
      butterCatalog.components.get('TextField')?.schema.safeParse({
        name: 'reason',
        label: 'Reason',
        value: { path: '/values/reason' },
        hint: 'Why now?',
        required: true,
        maxLength: 20,
      }).success
    ).toBe(true)
  })

  it('does not import extra components or executable catalog functions', () => {
    expect([...butterCatalog.components.keys()].sort()).toEqual(
      [
        'Button',
        'Card',
        'ChoicePicker',
        'Column',
        'Divider',
        'KeyValue',
        'Row',
        'Status',
        'Text',
        'TextField',
      ].sort()
    )
    expect(butterCatalog.functions.size).toBe(0)
  })
})
