/** Return a shorter display form of an IP when the runtime can normalize it. */
export function compressIp(ip) {
  const trimmed = ip?.trim()
  if (!trimmed) {
    return ''
  }
  if (!trimmed.includes(':')) {
    return trimmed
  }
  try {
    return new URL(`http://[${trimmed}]/`).hostname
  } catch {
    return trimmed
  }
}

/** Display label for audit log IPs; `full` is set when it differs from `display`. */
export function formatIpDisplay(ip) {
  const full = ip?.trim() || ''
  if (!full) {
    return { display: '—', full: '' }
  }
  const display = compressIp(full)
  if (display && display !== full) {
    return { display, full }
  }
  return { display: full, full: '' }
}
