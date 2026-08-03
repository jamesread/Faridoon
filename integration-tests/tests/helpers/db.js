import mysql from 'mysql2/promise'

function dbConfig() {
  return {
    host: process.env.DB_HOST || '127.0.0.1',
    port: Number(process.env.DB_PORT || 3306),
    user: process.env.DB_USER || process.env.DB_USERNAME || 'user',
    password: process.env.DB_PASS || process.env.DB_PASSWORD || 'password',
    database: process.env.DB_NAME || process.env.DB_DATABASE || 'faridoon',
  }
}

export async function withDb(fn) {
  const conn = await mysql.createConnection(dbConfig())
  try {
    return await fn(conn)
  } finally {
    await conn.end()
  }
}

/** Insert a quote row; returns insert id. approval 1 = published, 0 = pending. */
export async function insertQuote(conn, { content, approval }) {
  const [result] = await conn.execute(
    `INSERT INTO quotes (content, approval, created, submitted_by_username)
     VALUES (?, ?, NOW(), ?)`,
    [content, approval, approval === 1 ? 'search-test' : 'Guest'],
  )
  return result.insertId
}

export async function deleteQuotesByIds(conn, ids) {
  if (!ids.length) {
    return
  }
  const placeholders = ids.map(() => '?').join(',')
  await conn.execute(`DELETE FROM quotes WHERE id IN (${placeholders})`, ids)
}
