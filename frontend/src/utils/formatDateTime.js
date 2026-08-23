function parseServerDateTime(value) {
  if (!value) {
    return null
  }
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const date = new Date(normalized)
  if (Number.isNaN(date.getTime())) {
    return null
  }
  return date
}

/** Compact label for tables, e.g. "Aug 21, 11:24 AM". */
export function formatDateTimeShort(value, locale = undefined) {
  const date = parseServerDateTime(value)
  if (!date) {
    return value || '—'
  }
  return new Intl.DateTimeFormat(locale, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date)
}

/** Full label for detail views; falls back to the raw server value. */
export function formatDateTimeFull(value) {
  if (!value) {
    return '—'
  }
  return value
}
