import { describe, expect, it } from 'vitest'
import { apiKeyForSave } from './api-key'

describe('apiKeyForSave', () => {
  it('keeps the stored key when the field is empty', () => {
    expect(apiKeyForSave({ apiKey: '  ', clearKey: false })).toBeUndefined()
  })

  it('sets or rotates the key when one is typed', () => {
    expect(apiKeyForSave({ apiKey: ' m0sk_new ', clearKey: false })).toBe(
      'm0sk_new'
    )
  })

  it('clears the key when asked, even if one is typed', () => {
    expect(apiKeyForSave({ apiKey: 'm0sk_new', clearKey: true })).toBe('')
  })
})
