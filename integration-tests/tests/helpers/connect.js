import assert from 'assert'

const base = () => process.env.FARIDOON_BASE_URL || 'http://localhost:18080'

/**
 * Call a Faridoon ConnectRPC unary method with the JSON protocol.
 * @param {string} method e.g. "ListQuotes"
 * @param {object} body
 * @param {{ cookie?: string }} [opts]
 */
export async function rpc(method, body = {}, opts = {}) {
  const headers = {
    'Content-Type': 'application/json',
  }
  if (opts.cookie) {
    headers.Cookie = opts.cookie
  }
  const res = await fetch(`${base()}/faridoon.v1.FaridoonService/${method}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  })
  const text = await res.text()
  let data
  try {
    data = text ? JSON.parse(text) : {}
  } catch {
    data = { raw: text }
  }
  const setCookie = res.headers.getSetCookie?.() || []
  return { status: res.status, ok: res.ok, data, setCookie, headers: res.headers }
}

export function assertOk(result, label = 'rpc') {
  assert.ok(result.ok, `${label} HTTP ${result.status}: ${JSON.stringify(result.data)}`)
  assert.ok(!result.data.code, `${label} Connect error: ${JSON.stringify(result.data)}`)
}

export function cookieHeaderFromSetCookie(setCookie) {
  if (!setCookie || !setCookie.length) {
    return ''
  }
  return setCookie.map((c) => c.split(';')[0]).join('; ')
}
