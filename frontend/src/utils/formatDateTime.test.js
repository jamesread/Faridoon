import { describe, expect, it } from 'vitest'
import { formatDateTimeFull, formatDateTimeShort } from './formatDateTime.js'

describe('formatDateTimeShort', () => {
  it('formats server timestamps compactly', () => {
    expect(formatDateTimeShort('2026-08-21 11:24:56', 'en-US')).toBe('Aug 21, 11:24 AM')
  })

  it('returns em dash for empty input', () => {
    expect(formatDateTimeShort('')).toBe('—')
  })
})

describe('formatDateTimeFull', () => {
  it('returns the raw server timestamp', () => {
    expect(formatDateTimeFull('2026-08-21 11:24:56')).toBe('2026-08-21 11:24:56')
  })

  it('returns em dash for empty input', () => {
    expect(formatDateTimeFull('')).toBe('—')
  })
})
