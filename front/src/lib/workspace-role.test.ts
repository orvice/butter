import { describe, expect, it } from 'vitest'
import { canManageWorkspace } from './workspace-role'

describe('canManageWorkspace', () => {
  const members = [
    { userId: 'owner-1', role: 'owner' },
    { userId: 'admin-1', role: 'admin' },
    { userId: 'member-1', role: 'member' },
  ]

  it('grants global admins regardless of membership', () => {
    expect(canManageWorkspace(true, 'stranger', undefined)).toBe(true)
  })

  it('grants workspace owners and admins only', () => {
    expect(canManageWorkspace(false, 'owner-1', members)).toBe(true)
    expect(canManageWorkspace(false, 'admin-1', members)).toBe(true)
    expect(canManageWorkspace(false, 'member-1', members)).toBe(false)
    expect(canManageWorkspace(false, 'stranger', members)).toBe(false)
  })

  it('denies while identity or membership is unknown', () => {
    expect(canManageWorkspace(false, undefined, members)).toBe(false)
    expect(canManageWorkspace(false, 'owner-1', undefined)).toBe(false)
  })
})
