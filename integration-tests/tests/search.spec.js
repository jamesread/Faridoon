import assert from 'assert'
import { rpc, assertOk } from './helpers/connect.js'
import { withDb, insertQuote, deleteQuotesByIds } from './helpers/db.js'

describe('Quote search (ListQuotes query)', function () {
  this.timeout(30000)

  const runId = `${Date.now()}`
  const approvedToken = `srchapprv${runId}`
  const pendingToken = `srchpend${runId}`
  // Two-character ASCII needle forces the LIKE fallback (below InnoDB ft min token length).
  const shortNeedle = 'z@'
  const approvedContent = `Alpha ${approvedToken} line with needle ${shortNeedle} omega`
  const pendingContent = `Secret ${pendingToken} must stay unpublished`
  const decoyContent = `Unrelated decoy content ${runId} without search tokens`

  /** @type {number[]} */
  let quoteIds = []

  before(async function () {
    await withDb(async (conn) => {
      quoteIds.push(await insertQuote(conn, { content: approvedContent, approval: 1 }))
      quoteIds.push(await insertQuote(conn, { content: pendingContent, approval: 0 }))
      quoteIds.push(await insertQuote(conn, { content: decoyContent, approval: 1 }))
    })
  })

  after(async function () {
    await withDb(async (conn) => {
      await deleteQuotesByIds(conn, quoteIds)
    })
  })

  async function search(query) {
    const result = await rpc('ListQuotes', { page: 1, order: 'latest', query })
    assertOk(result, 'ListQuotes')
    return result.data
  }

  it('returns approved quotes matching a full-text token', async function () {
    const data = await search(approvedToken)
    const ids = (data.quotes || []).map((q) => q.id)
    assert.ok(ids.includes(quoteIds[0]), `expected approved id ${quoteIds[0]} in ${JSON.stringify(ids)}`)
    assert.ok(
      (data.quotes || []).some((q) => (q.content || '').includes(approvedToken)),
      'approved quote content should appear in results',
    )
  })

  it('does not reveal pending (unapproved) quotes', async function () {
    const data = await search(pendingToken)
    const quotes = data.quotes || []
    const ids = quotes.map((q) => q.id)
    assert.ok(!ids.includes(quoteIds[1]), `pending id ${quoteIds[1]} must not appear in ${JSON.stringify(ids)}`)
    for (const q of quotes) {
      assert.ok(!(q.content || '').includes(pendingToken), `pending content leaked in quote #${q.id}`)
    }
    assert.equal(data.total || 0, 0, 'total should be 0 when only pending matches')
  })

  it('finds approved quotes via short LIKE fallback', async function () {
    // Tokens shorter than InnoDB ft min length use LIKE; shortNeedle is embedded in approved content.
    const data = await search(shortNeedle)
    const ids = (data.quotes || []).map((q) => q.id)
    assert.ok(ids.includes(quoteIds[0]), `expected approved id ${quoteIds[0]} for short query in ${JSON.stringify(ids)}`)
  })

  it('returns no quotes for a non-matching query', async function () {
    const data = await search(`nomatch${runId}zzz`)
    assert.equal((data.quotes || []).length, 0)
    assert.equal(data.total || 0, 0)
  })

  it('list without query still excludes pending quotes', async function () {
    const data = await search('')
    const ids = (data.quotes || []).map((q) => q.id)
    assert.ok(!ids.includes(quoteIds[1]), 'unfiltered list must not include pending quote')
  })
})
