import { describe, expect, it } from 'vitest'
import { compressIp, formatIpDisplay } from './formatIp.js'

describe('compressIp', () => {
  it('returns IPv4 unchanged', () => {
    expect(compressIp('127.0.0.1')).toBe('127.0.0.1')
  })

  it('compresses expanded IPv6 when supported by URL parsing', () => {
    const compressed = compressIp('2001:0db8:85a3:0000:0000:8a2e:0370:7334')
    expect(compressed).toContain('2001')
    expect(compressed.length).toBeLessThan('2001:0db8:85a3:0000:0000:8a2e:0370:7334'.length)
  })
})

describe('formatIpDisplay', () => {
  it('returns em dash for empty input', () => {
    expect(formatIpDisplay('')).toEqual({ display: '—', full: '' })
  })

  it('keeps full value when compression is unavailable', () => {
    expect(formatIpDisplay('127.0.0.1')).toEqual({ display: '127.0.0.1', full: '' })
  })
})
