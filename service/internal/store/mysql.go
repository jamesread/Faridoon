package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// MySQL keyword/reserved identifiers used as column/table names in Faridoon's schema.
// Quote opportunistically so MySQL 8+ reserved-word changes do not break queries.
const (
	colGroup    = "`group`"
	colKey      = "`key`"
	colUser     = "`user`"
	colEvent    = "`event`"
	colPassword = "`password`"
	colAction   = "`action`"
	tableGroups = "`groups`"
	tableLogs   = "`logs`"
)

type QuoteRow struct {
	Content             string
	Created             string
	SyntaxHighlighting  string
	SubmittedByUsername string
	ID                  int
	VoteCount           int
	Approval            int
	SubmittedByUserID   int
	Approved            bool
}

type UserRow struct {
	Username   string
	Password   string
	GroupTitle string
	ID         int
	GroupID    int
}

type GroupRow struct {
	Title string
	ID    int
}

type PermissionRow struct {
	Key         string
	Description string
	ID          int
}

type GroupPermissionRow struct {
	Key          string
	Description  string
	PermissionID int
}

type WebhookTargetRow struct {
	URL     string
	Secret  string
	Created string
	Updated string
	Events  []string
	ID      int
	Enabled bool
}

type LogEntry struct {
	Created       string
	ActorUsername string
	Action        string
	EntityType    string
	Detail        string
	IP            string
	ID            int
	ActorUserID   int
	EntityID      int
}

type HeaderLinkRow struct {
	Title        string
	URL          string
	Created      string
	Updated      string
	ID           int
	SortOrder    int
	Enabled      bool
	OpenInNewTab bool
}

type Store interface {
	LatestMigration(ctx context.Context) (string, error)
	HasMigration(ctx context.Context, id string) (bool, error)
	FindQuote(ctx context.Context, id int) (*QuoteRow, error)
	FindQuoteRaw(ctx context.Context, id int) (*QuoteRow, error)
	ListApproved(ctx context.Context, order string, page, perPage int, query string) ([]QuoteRow, int, error)
	ListPending(ctx context.Context) ([]QuoteRow, error)
	CountPending(ctx context.Context) (int, error)
	CountApproved(ctx context.Context) (int, error)
	CountUsersWithPrivilege(ctx context.Context, key string) (int, error)
	CountUsersWithPrivilegeExcludingGroup(ctx context.Context, key string, groupID int) (int, error)
	FindPermission(ctx context.Context, id int) (*PermissionRow, error)
	CreateQuote(ctx context.Context, content string, approval int, syntax string, submittedByUserID int, submittedByUsername string) (int, error)
	UpdateQuote(ctx context.Context, id int, content, syntax string) error
	ApproveQuote(ctx context.Context, id int) error
	DeleteQuote(ctx context.Context, id int) error
	VoteSum(ctx context.Context, quoteID int) (int, error)
	CastVote(ctx context.Context, quoteID, userID, delta int) error
	UserCount(ctx context.Context) (int, error)
	FindUserByUsername(ctx context.Context, username string) (*UserRow, error)
	FindUser(ctx context.Context, id int) (*UserRow, error)
	CreateUser(ctx context.Context, username, passwordHash string, groupID int) (int, error)
	UpdatePassword(ctx context.Context, userID int, passwordHash string) error
	UpdateUserGroup(ctx context.Context, userID, groupID int) error
	DeleteUser(ctx context.Context, id int) error
	ListUsers(ctx context.Context) ([]UserRow, error)
	UserPrivileges(ctx context.Context, userID, groupID int) ([]string, error)
	ListGroups(ctx context.Context) ([]GroupRow, error)
	FindGroup(ctx context.Context, id int) (*GroupRow, error)
	GroupPermissions(ctx context.Context, groupID int) ([]GroupPermissionRow, error)
	GroupsWithPermissions(ctx context.Context) ([]GroupRow, map[int][]GroupPermissionRow, error)
	CreateGroup(ctx context.Context, title string) (int, error)
	DeleteGroup(ctx context.Context, id int) error
	ListPermissions(ctx context.Context) ([]PermissionRow, error)
	GrantPermission(ctx context.Context, groupID, permissionID int) error
	RevokePermission(ctx context.Context, groupID, permissionID int) error
	ListWebhookTargets(ctx context.Context) ([]WebhookTargetRow, error)
	FindWebhookTarget(ctx context.Context, id int) (*WebhookTargetRow, error)
	EnabledTargetsForEvent(ctx context.Context, event string) ([]WebhookTargetRow, error)
	CreateWebhookTarget(ctx context.Context, url, secret string, events []string, enabled bool) (int, error)
	UpdateWebhookTarget(ctx context.Context, id int, url, secret string, events []string, enabled bool, clearSecret bool) error
	DeleteWebhookTarget(ctx context.Context, id int) error
	InsertLog(ctx context.Context, entry LogEntry) error
	ListLogs(ctx context.Context, page, pageSize int) ([]LogEntry, int, error)
	ListHeaderLinks(ctx context.Context) ([]HeaderLinkRow, error)
	ListEnabledHeaderLinks(ctx context.Context) ([]HeaderLinkRow, error)
	FindHeaderLink(ctx context.Context, id int) (*HeaderLinkRow, error)
	CreateHeaderLink(ctx context.Context, title, url string, sortOrder int, enabled, openInNewTab bool) (int, error)
	UpdateHeaderLink(ctx context.Context, id int, title, url string, sortOrder int, enabled, openInNewTab bool) error
	DeleteHeaderLink(ctx context.Context, id int) error
	ListCvars(ctx context.Context) ([]CvarRow, error)
	FindCvar(ctx context.Context, key string) (*CvarRow, error)
	InsertCvarIfMissing(ctx context.Context, row CvarRow) error
	UpdateCvar(ctx context.Context, key string, valueInt int, valueString string) error
}

type CvarRow struct {
	Key         string
	MainType    string
	Title       string
	Description string
	Category    string
	ValueString string
	Ordinal     int
	ValueInt    int
}

type MySQL struct {
	db *sql.DB
}

func NewMySQL(db *sql.DB) *MySQL {
	return &MySQL{db: db}
}

func (m *MySQL) LatestMigration(ctx context.Context) (string, error) {
	var id sql.NullString
	err := m.db.QueryRowContext(ctx,
		`SELECT id FROM migrations ORDER BY applied_at DESC, id DESC LIMIT 1`).Scan(&id)
	if err != nil {
		return "", err
	}
	if !id.Valid {
		return "", nil
	}
	return id.String, nil
}

func (m *MySQL) HasMigration(ctx context.Context, id string) (bool, error) {
	var n int
	err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM migrations WHERE id = ?`, id).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func quoteSelectSQL() string {
	return `
SELECT q.id, q.content, COALESCE(DATE_FORMAT(q.created, '%Y-%m-%d %H:%i:%s'), ''),
  q.approval, COALESCE(q.syntaxHighlighting, ''), COALESCE(v.voteCount, 0),
  COALESCE(q.submitted_by_user_id, 0), COALESCE(q.submitted_by_username, '')
FROM quotes q
LEFT JOIN (
  SELECT quote, COALESCE(SUM(delta), 0) AS voteCount FROM votes GROUP BY quote
) v ON v.quote = q.id`
}

func scanQuote(s interface{ Scan(...any) error }) (*QuoteRow, error) {
	var q QuoteRow
	var approval int
	if err := s.Scan(&q.ID, &q.Content, &q.Created, &approval, &q.SyntaxHighlighting, &q.VoteCount,
		&q.SubmittedByUserID, &q.SubmittedByUsername); err != nil {
		return nil, err
	}
	q.Approval = approval
	q.Approved = approval == 1
	return &q, nil
}

func scanQuotes(rows *sql.Rows) ([]QuoteRow, error) {
	var out []QuoteRow
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *q)
	}
	return out, nil
}

func approvedOrderSQL(order string) string {
	switch order {
	case "random":
		return "ORDER BY RAND()"
	case "rank":
		return "ORDER BY voteCount DESC, q.created DESC"
	default:
		return "ORDER BY q.created DESC"
	}
}

// fulltextMinTokenLen matches InnoDB's default innodb_ft_min_token_size.
const fulltextMinTokenLen = 3

func buildBooleanQuery(raw string) string {
	fields := strings.Fields(raw)
	var parts []string
	for _, field := range fields {
		token := sanitizeFulltextToken(field)
		if utf8.RuneCountInString(token) < fulltextMinTokenLen {
			continue
		}
		parts = append(parts, "+"+token+"*")
	}
	return strings.Join(parts, " ")
}

func sanitizeFulltextToken(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '+', '-', '>', '<', '(', ')', '~', '*', '"', '@':
			continue
		default:
			if unicode.IsSpace(r) {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

// likeEscapeChar is used in LIKE ... ESCAPE so user %/_ cannot act as wildcards.
// Avoids backslash, which is awkward in Go/SQL string literals.
const likeEscapeChar = "|"

func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(
		likeEscapeChar, likeEscapeChar+likeEscapeChar,
		"%", likeEscapeChar+"%",
		"_", likeEscapeChar+"_",
	)
	return replacer.Replace(s)
}

func (m *MySQL) FindQuote(ctx context.Context, id int) (*QuoteRow, error) {
	row := m.db.QueryRowContext(ctx, quoteSelectSQL()+" WHERE q.id = ?", id)
	q, err := scanQuote(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return q, err
}

func (m *MySQL) FindQuoteRaw(ctx context.Context, id int) (*QuoteRow, error) {
	row := m.db.QueryRowContext(ctx,
		`SELECT id, content, COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), ''), approval,
		 COALESCE(syntaxHighlighting, ''), 0, COALESCE(submitted_by_user_id, 0), COALESCE(submitted_by_username, '')
		 FROM quotes WHERE id = ?`, id)
	q, err := scanQuote(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return q, err
}

func (m *MySQL) ListApproved(ctx context.Context, order string, page, perPage int, query string) ([]QuoteRow, int, error) {
	query = strings.TrimSpace(query)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 1
	}
	offset := (page - 1) * perPage

	if query == "" {
		return m.listApprovedUnfiltered(ctx, order, perPage, offset)
	}

	booleanQuery := buildBooleanQuery(query)
	if booleanQuery != "" {
		return m.listApprovedFulltext(ctx, order, perPage, offset, booleanQuery)
	}
	return m.listApprovedLike(ctx, order, perPage, offset, query)
}

func (m *MySQL) listApprovedUnfiltered(ctx context.Context, order string, perPage, offset int) ([]QuoteRow, int, error) {
	var total int
	if err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quotes WHERE approval = 1`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := m.db.QueryContext(ctx, quoteSelectSQL()+" WHERE "+approvedQuoteFilterSQL()+" "+approvedOrderSQL(order)+" LIMIT ? OFFSET ?", perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out, err := scanQuotes(rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// approvedQuoteFilterSQL is the shared predicate for public listing and search.
// Pending (unapproved) quotes must never appear in ListApproved results.
func approvedQuoteFilterSQL() string {
	return "q.approval = 1"
}

func (m *MySQL) listApprovedFulltext(ctx context.Context, order string, perPage, offset int, booleanQuery string) ([]QuoteRow, int, error) {
	var total int
	if err := m.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM quotes WHERE approval = 1 AND MATCH(content) AGAINST (? IN BOOLEAN MODE)`,
		booleanQuery,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	where := approvedQuoteFilterSQL() + " AND MATCH(q.content) AGAINST (? IN BOOLEAN MODE)"
	orderSQL, args := fulltextOrderAndArgs(order, booleanQuery)
	args = append(args, perPage, offset)
	rows, err := m.db.QueryContext(ctx,
		quoteSelectSQL()+" WHERE "+where+" "+orderSQL+" LIMIT ? OFFSET ?",
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out, err := scanQuotes(rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func fulltextOrderAndArgs(order, booleanQuery string) (string, []any) {
	args := []any{booleanQuery}
	if order == "" || order == "latest" {
		return "ORDER BY MATCH(q.content) AGAINST (? IN BOOLEAN MODE) DESC, q.created DESC",
			append(args, booleanQuery)
	}
	return approvedOrderSQL(order), args
}

func (m *MySQL) listApprovedLike(ctx context.Context, order string, perPage, offset int, query string) ([]QuoteRow, int, error) {
	pattern := "%" + escapeLikePattern(query) + "%"
	var total int
	if err := m.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM quotes WHERE approval = 1 AND content LIKE ? ESCAPE '`+likeEscapeChar+`'`,
		pattern,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	where := approvedQuoteFilterSQL() + ` AND q.content LIKE ? ESCAPE '` + likeEscapeChar + `'`
	rows, err := m.db.QueryContext(ctx,
		quoteSelectSQL()+" WHERE "+where+" "+approvedOrderSQL(order)+" LIMIT ? OFFSET ?",
		pattern, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out, err := scanQuotes(rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (m *MySQL) ListPending(ctx context.Context) ([]QuoteRow, error) {
	rows, err := m.db.QueryContext(ctx, quoteSelectSQL()+" WHERE q.approval = 0 ORDER BY q.created DESC")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanQuotes(rows)
}

func (m *MySQL) CountPending(ctx context.Context) (int, error) {
	var n int
	err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quotes WHERE approval = 0`).Scan(&n)
	return n, err
}

func (m *MySQL) CountApproved(ctx context.Context) (int, error) {
	var n int
	err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quotes WHERE approval = 1`).Scan(&n)
	return n, err
}

func (m *MySQL) CountUsersWithPrivilege(ctx context.Context, key string) (int, error) {
	var n int
	err := m.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM (
  SELECT u.id AS id FROM users u
  JOIN privileges_g pg ON pg.`+colGroup+` = u.`+colGroup+`
  JOIN permissions p ON p.id = pg.permission AND p.`+colKey+` = ?
  UNION
  SELECT pu.`+colUser+` AS id FROM privileges_u pu
  JOIN permissions p ON p.id = pu.permission AND p.`+colKey+` = ?
) t`, key, key).Scan(&n)
	return n, err
}

func (m *MySQL) CountUsersWithPrivilegeExcludingGroup(ctx context.Context, key string, groupID int) (int, error) {
	var n int
	err := m.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM (
  SELECT u.id AS id FROM users u
  JOIN privileges_g pg ON pg.`+colGroup+` = u.`+colGroup+`
  JOIN permissions p ON p.id = pg.permission AND p.`+colKey+` = ?
  WHERE u.`+colGroup+` <> ?
  UNION
  SELECT pu.`+colUser+` AS id FROM privileges_u pu
  JOIN permissions p ON p.id = pu.permission AND p.`+colKey+` = ?
) t`, key, groupID, key).Scan(&n)
	return n, err
}

func (m *MySQL) FindPermission(ctx context.Context, id int) (*PermissionRow, error) {
	row := m.db.QueryRowContext(ctx,
		"SELECT id, "+colKey+", COALESCE(description, '') FROM permissions WHERE id = ?", id)
	var p PermissionRow
	if err := row.Scan(&p.ID, &p.Key, &p.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (m *MySQL) CreateQuote(ctx context.Context, content string, approval int, syntax string, submittedByUserID int, submittedByUsername string) (int, error) {
	var userID any
	if submittedByUserID > 0 {
		userID = submittedByUserID
	}
	res, err := m.db.ExecContext(ctx,
		`INSERT INTO quotes (content, approval, syntaxHighlighting, created, submitted_by_user_id, submitted_by_username)
		 VALUES (?, ?, NULLIF(?, ''), NOW(), ?, NULLIF(?, ''))`,
		content, approval, syntax, userID, submittedByUsername)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) UpdateQuote(ctx context.Context, id int, content, syntax string) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE quotes SET content = ?, syntaxHighlighting = NULLIF(?, '') WHERE id = ?`,
		content, syntax, id)
	return err
}

func (m *MySQL) ApproveQuote(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `UPDATE quotes SET approval = 1 WHERE id = ?`, id)
	return err
}

func (m *MySQL) DeleteQuote(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM quotes WHERE id = ? LIMIT 1`, id)
	return err
}

func (m *MySQL) VoteSum(ctx context.Context, quoteID int) (int, error) {
	var n sql.NullInt64
	err := m.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(delta), 0) FROM votes WHERE quote = ?`, quoteID).Scan(&n)
	return int(n.Int64), err
}

func (m *MySQL) CastVote(ctx context.Context, quoteID, userID, delta int) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "DELETE FROM votes WHERE quote = ? AND "+colUser+" = ?", quoteID, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO votes (quote, "+colUser+", delta) VALUES (?, ?, ?)", quoteID, userID, delta); err != nil {
		return err
	}
	return tx.Commit()
}

func (m *MySQL) UserCount(ctx context.Context) (int, error) {
	var n int
	err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func userSelectSQL() string {
	return "SELECT u.id, u.username, u." + colPassword + ", COALESCE(u." + colGroup + ", 0), COALESCE(g.title, '') FROM users u LEFT JOIN " + tableGroups + " g ON u." + colGroup + " = g.id"
}

func scanUser(s interface{ Scan(...any) error }) (*UserRow, error) {
	var u UserRow
	var groupTitle sql.NullString
	if err := s.Scan(&u.ID, &u.Username, &u.Password, &u.GroupID, &groupTitle); err != nil {
		return nil, err
	}
	u.GroupTitle = groupTitle.String
	return &u, nil
}

func scanUsers(rows *sql.Rows) ([]UserRow, error) {
	var out []UserRow
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, nil
}

func (m *MySQL) FindUserByUsername(ctx context.Context, username string) (*UserRow, error) {
	row := m.db.QueryRowContext(ctx, userSelectSQL()+" WHERE u.username = ?", username)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (m *MySQL) FindUser(ctx context.Context, id int) (*UserRow, error) {
	row := m.db.QueryRowContext(ctx, userSelectSQL()+" WHERE u.id = ?", id)
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (m *MySQL) CreateUser(ctx context.Context, username, passwordHash string, groupID int) (int, error) {
	res, err := m.db.ExecContext(ctx,
		"INSERT INTO users (username, "+colPassword+", "+colGroup+", registered) VALUES (?, ?, ?, NOW())",
		username, passwordHash, groupID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) UpdatePassword(ctx context.Context, userID int, passwordHash string) error {
	_, err := m.db.ExecContext(ctx, "UPDATE users SET "+colPassword+" = ? WHERE id = ?", passwordHash, userID)
	return err
}

func (m *MySQL) UpdateUserGroup(ctx context.Context, userID, groupID int) error {
	_, err := m.db.ExecContext(ctx, "UPDATE users SET "+colGroup+" = ? WHERE id = ?", groupID, userID)
	return err
}

func (m *MySQL) DeleteUser(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (m *MySQL) ListUsers(ctx context.Context) ([]UserRow, error) {
	rows, err := m.db.QueryContext(ctx, userSelectSQL()+" ORDER BY u.id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanUsers(rows)
}

func (m *MySQL) loadPrivKeys(ctx context.Context, query string, args ...any) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var keys []string
	for rows.Next() {
		var k string
		if scanErr := rows.Scan(&k); scanErr != nil {
			return nil, scanErr
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *MySQL) UserPrivileges(ctx context.Context, userID, groupID int) ([]string, error) {
	q := "SELECT p." + colKey + " FROM privileges_u pu JOIN permissions p ON p.id = pu.permission WHERE pu." + colUser + " = ?" +
		" UNION SELECT p." + colKey + " FROM privileges_g pg JOIN permissions p ON p.id = pg.permission WHERE pg." + colGroup + " = ?"
	return m.loadPrivKeys(ctx, q, userID, groupID)
}

func (m *MySQL) ListGroups(ctx context.Context) ([]GroupRow, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT id, title FROM "+tableGroups+" ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []GroupRow
	for rows.Next() {
		var g GroupRow
		if err := rows.Scan(&g.ID, &g.Title); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (m *MySQL) FindGroup(ctx context.Context, id int) (*GroupRow, error) {
	row := m.db.QueryRowContext(ctx, "SELECT id, title FROM "+tableGroups+" WHERE id = ?", id)
	var g GroupRow
	if err := row.Scan(&g.ID, &g.Title); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (m *MySQL) GroupPermissions(ctx context.Context, groupID int) ([]GroupPermissionRow, error) {
	return m.loadGroupPerms(ctx, groupID)
}

func (m *MySQL) loadGroupPerms(ctx context.Context, groupID int) ([]GroupPermissionRow, error) {
	q := "SELECT gp.permission, COALESCE(p." + colKey + ", ''), COALESCE(p.description, '') FROM privileges_g gp LEFT JOIN permissions p ON gp.permission = p.id WHERE gp." + colGroup + " = ?"
	rows, err := m.db.QueryContext(ctx, q, groupID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []GroupPermissionRow
	for rows.Next() {
		var gp GroupPermissionRow
		if scanErr := rows.Scan(&gp.PermissionID, &gp.Key, &gp.Description); scanErr != nil {
			return nil, scanErr
		}
		out = append(out, gp)
	}
	return out, nil
}

func (m *MySQL) GroupsWithPermissions(ctx context.Context) ([]GroupRow, map[int][]GroupPermissionRow, error) {
	groups, err := m.ListGroups(ctx)
	if err != nil {
		return nil, nil, err
	}
	perms := make(map[int][]GroupPermissionRow, len(groups))
	for _, g := range groups {
		perms[g.ID] = nil
	}
	if loadErr := m.loadAllGroupPermissions(ctx, perms); loadErr != nil {
		return nil, nil, loadErr
	}
	return groups, perms, nil
}

func (m *MySQL) loadAllGroupPermissions(ctx context.Context, perms map[int][]GroupPermissionRow) error {
	q := "SELECT gp." + colGroup + ", gp.permission, COALESCE(p." + colKey + ", ''), COALESCE(p.description, '') FROM privileges_g gp LEFT JOIN permissions p ON gp.permission = p.id"
	rows, err := m.db.QueryContext(ctx, q)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var groupID int
		var gp GroupPermissionRow
		if scanErr := rows.Scan(&groupID, &gp.PermissionID, &gp.Key, &gp.Description); scanErr != nil {
			return scanErr
		}
		perms[groupID] = append(perms[groupID], gp)
	}
	return nil
}

func (m *MySQL) CreateGroup(ctx context.Context, title string) (int, error) {
	res, err := m.db.ExecContext(ctx, "INSERT INTO "+tableGroups+" (title) VALUES (?)", title)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) DeleteGroup(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, "DELETE FROM "+tableGroups+" WHERE id = ? LIMIT 1", id)
	return err
}

func (m *MySQL) ListPermissions(ctx context.Context) ([]PermissionRow, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT id, "+colKey+", COALESCE(description, '') FROM permissions ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []PermissionRow
	for rows.Next() {
		var p PermissionRow
		if err := rows.Scan(&p.ID, &p.Key, &p.Description); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (m *MySQL) GrantPermission(ctx context.Context, groupID, permissionID int) error {
	var n int
	err := m.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM privileges_g WHERE "+colGroup+" = ? AND permission = ?", groupID, permissionID).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err = m.db.ExecContext(ctx, "INSERT INTO privileges_g ("+colGroup+", permission) VALUES (?, ?)", groupID, permissionID)
	return err
}

func (m *MySQL) RevokePermission(ctx context.Context, groupID, permissionID int) error {
	_, err := m.db.ExecContext(ctx,
		"DELETE FROM privileges_g WHERE "+colGroup+" = ? AND permission = ?", groupID, permissionID)
	return err
}

func webhookTargetSelectSQL() string {
	return "SELECT id, url, secret, enabled," +
		" COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), '')," +
		" COALESCE(DATE_FORMAT(updated, '%Y-%m-%d %H:%i:%s'), '')" +
		" FROM webhook_targets"
}

func (m *MySQL) ListWebhookTargets(ctx context.Context) ([]WebhookTargetRow, error) {
	rows, err := m.db.QueryContext(ctx, webhookTargetSelectSQL()+" ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	targets, err := scanWebhookTargets(rows)
	if err != nil {
		return nil, err
	}
	for i := range targets {
		events, loadErr := m.loadWebhookEvents(ctx, targets[i].ID)
		if loadErr != nil {
			return nil, loadErr
		}
		targets[i].Events = events
	}
	return targets, nil
}

func (m *MySQL) FindWebhookTarget(ctx context.Context, id int) (*WebhookTargetRow, error) {
	row := m.db.QueryRowContext(ctx, webhookTargetSelectSQL()+" WHERE id = ?", id)
	w, err := scanWebhookTarget(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	events, loadErr := m.loadWebhookEvents(ctx, w.ID)
	if loadErr != nil {
		return nil, loadErr
	}
	w.Events = events
	return w, nil
}

func (m *MySQL) EnabledTargetsForEvent(ctx context.Context, event string) ([]WebhookTargetRow, error) {
	q := "SELECT t.id, t.url, t.secret, t.enabled," +
		" COALESCE(DATE_FORMAT(t.created, '%Y-%m-%d %H:%i:%s'), '')," +
		" COALESCE(DATE_FORMAT(t.updated, '%Y-%m-%d %H:%i:%s'), '')" +
		" FROM webhook_targets t" +
		" INNER JOIN webhook_events e ON e.webhook_target_id = t.id" +
		" WHERE e." + colEvent + " = ? AND t.enabled = 1"
	rows, err := m.db.QueryContext(ctx, q, event)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWebhookTargets(rows)
}

func (m *MySQL) CreateWebhookTarget(ctx context.Context, url, secret string, events []string, enabled bool) (int, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	id, err := insertWebhookTargetTx(ctx, tx, url, secret, events, enabled)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func insertWebhookTargetTx(ctx context.Context, tx *sql.Tx, url, secret string, events []string, enabled bool) (int, error) {
	now := time.Now()
	res, err := tx.ExecContext(ctx,
		`INSERT INTO webhook_targets (url, secret, enabled, created, updated) VALUES (?, ?, ?, ?, ?)`,
		url, secret, boolToTinyInt(enabled), now, now)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := insertWebhookEventsTx(ctx, tx, int(lid), events); err != nil {
		return 0, err
	}
	return int(lid), nil
}

func (m *MySQL) UpdateWebhookTarget(ctx context.Context, id int, url, secret string, events []string, enabled bool, clearSecret bool) error {
	cur, err := m.FindWebhookTarget(ctx, id)
	if err != nil || cur == nil {
		return err
	}
	applyWebhookTargetPatch(cur, url, secret, clearSecret)
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := writeWebhookTargetTx(ctx, tx, id, cur.URL, cur.Secret, enabled, events); err != nil {
		return err
	}
	return tx.Commit()
}

func applyWebhookTargetPatch(cur *WebhookTargetRow, url, secret string, clearSecret bool) {
	if url != "" {
		cur.URL = url
	}
	if !clearSecret && secret != "" {
		cur.Secret = secret
	}
}

func writeWebhookTargetTx(ctx context.Context, tx *sql.Tx, id int, url, secret string, enabled bool, events []string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE webhook_targets SET url = ?, secret = ?, enabled = ?, updated = ? WHERE id = ? LIMIT 1`,
		url, secret, boolToTinyInt(enabled), time.Now(), id)
	if err != nil {
		return err
	}
	return replaceWebhookEventsTx(ctx, tx, id, events)
}

func replaceWebhookEventsTx(ctx context.Context, tx *sql.Tx, targetID int, events []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM webhook_events WHERE webhook_target_id = ?`, targetID); err != nil {
		return err
	}
	return insertWebhookEventsTx(ctx, tx, targetID, events)
}

func (m *MySQL) DeleteWebhookTarget(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM webhook_targets WHERE id = ? LIMIT 1`, id)
	return err
}

func (m *MySQL) loadWebhookEvents(ctx context.Context, targetID int) ([]string, error) {
	rows, err := m.db.QueryContext(ctx,
		"SELECT "+colEvent+" FROM webhook_events WHERE webhook_target_id = ? ORDER BY "+colEvent, targetID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var events []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func insertWebhookEventsTx(ctx context.Context, tx *sql.Tx, targetID int, events []string) error {
	for _, e := range events {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO webhook_events (webhook_target_id, "+colEvent+") VALUES (?, ?)", targetID, e); err != nil {
			return err
		}
	}
	return nil
}

func scanWebhookTarget(s interface{ Scan(...any) error }) (*WebhookTargetRow, error) {
	var w WebhookTargetRow
	var enabled int
	if err := s.Scan(&w.ID, &w.URL, &w.Secret, &enabled, &w.Created, &w.Updated); err != nil {
		return nil, err
	}
	w.Enabled = enabled == 1
	return &w, nil
}

func scanWebhookTargets(rows *sql.Rows) ([]WebhookTargetRow, error) {
	var out []WebhookTargetRow
	for rows.Next() {
		w, err := scanWebhookTarget(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	return out, nil
}

func (m *MySQL) InsertLog(ctx context.Context, entry LogEntry) error {
	created := time.Now()
	var actorID any
	if entry.ActorUserID > 0 {
		actorID = entry.ActorUserID
	}
	var entityID any
	if entry.EntityID > 0 {
		entityID = entry.EntityID
	}
	_, err := m.db.ExecContext(ctx,
		"INSERT INTO "+tableLogs+" (created, actor_user_id, actor_username, "+colAction+", entity_type, entity_id, detail, ip)"+
			" VALUES (?, ?, NULLIF(?, ''), ?, NULLIF(?, ''), ?, NULLIF(?, ''), NULLIF(?, ''))",
		created, actorID, entry.ActorUsername, entry.Action, entry.EntityType, entityID, entry.Detail, entry.IP)
	return err
}

func clampLogPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func scanLogRows(rows *sql.Rows) ([]LogEntry, error) {
	var out []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(&e.ID, &e.Created, &e.ActorUserID, &e.ActorUsername, &e.Action,
			&e.EntityType, &e.EntityID, &e.Detail, &e.IP); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (m *MySQL) ListLogs(ctx context.Context, page, pageSize int) ([]LogEntry, int, error) {
	page, pageSize = clampLogPage(page, pageSize)
	var total int
	if err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tableLogs).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := m.db.QueryContext(ctx,
		"SELECT id, COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), ''),"+
			" COALESCE(actor_user_id, 0), COALESCE(actor_username, ''), "+colAction+","+
			" COALESCE(entity_type, ''), COALESCE(entity_id, 0), COALESCE(detail, ''), COALESCE(ip, '')"+
			" FROM "+tableLogs+" ORDER BY id DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out, err := scanLogRows(rows)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func headerLinkSelectSQL() string {
	return `SELECT id, title, url, sort_order, enabled, open_in_new_tab,
		 COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), ''),
		 COALESCE(DATE_FORMAT(updated, '%Y-%m-%d %H:%i:%s'), '')
		 FROM header_links`
}

func scanHeaderLink(s interface{ Scan(...any) error }) (*HeaderLinkRow, error) {
	var row HeaderLinkRow
	var enabled, openInNewTab int
	if err := s.Scan(&row.ID, &row.Title, &row.URL, &row.SortOrder, &enabled, &openInNewTab, &row.Created, &row.Updated); err != nil {
		return nil, err
	}
	row.Enabled = enabled == 1
	row.OpenInNewTab = openInNewTab == 1
	return &row, nil
}

func scanHeaderLinks(rows *sql.Rows) ([]HeaderLinkRow, error) {
	var out []HeaderLinkRow
	for rows.Next() {
		row, err := scanHeaderLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, nil
}

func boolToTinyInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (m *MySQL) ListHeaderLinks(ctx context.Context) ([]HeaderLinkRow, error) {
	rows, err := m.db.QueryContext(ctx, headerLinkSelectSQL()+" ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanHeaderLinks(rows)
}

func (m *MySQL) ListEnabledHeaderLinks(ctx context.Context) ([]HeaderLinkRow, error) {
	rows, err := m.db.QueryContext(ctx, headerLinkSelectSQL()+" WHERE enabled = 1 ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanHeaderLinks(rows)
}

func (m *MySQL) FindHeaderLink(ctx context.Context, id int) (*HeaderLinkRow, error) {
	row := m.db.QueryRowContext(ctx, headerLinkSelectSQL()+" WHERE id = ?", id)
	link, err := scanHeaderLink(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return link, err
}

func (m *MySQL) CreateHeaderLink(ctx context.Context, title, url string, sortOrder int, enabled, openInNewTab bool) (int, error) {
	now := time.Now()
	res, err := m.db.ExecContext(ctx,
		`INSERT INTO header_links (title, url, sort_order, enabled, open_in_new_tab, created, updated)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		title, url, sortOrder, boolToTinyInt(enabled), boolToTinyInt(openInNewTab), now, now)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) UpdateHeaderLink(ctx context.Context, id int, title, url string, sortOrder int, enabled, openInNewTab bool) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE header_links SET title = ?, url = ?, sort_order = ?, enabled = ?, open_in_new_tab = ?, updated = ?
		 WHERE id = ? LIMIT 1`,
		title, url, sortOrder, boolToTinyInt(enabled), boolToTinyInt(openInNewTab), time.Now(), id)
	return err
}

func (m *MySQL) DeleteHeaderLink(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM header_links WHERE id = ? LIMIT 1`, id)
	return err
}

func cvarSelectSQL() string {
	return `SELECT cvar_key, COALESCE(cvar_value_int, 0), COALESCE(cvar_value_string, ''), cvar_main_type,
		COALESCE(cvar_title, ''), COALESCE(cvar_description, ''), COALESCE(cvar_category, ''), COALESCE(cvar_ordinal, 0)
		FROM cvars`
}

func scanCvar(s interface{ Scan(...any) error }) (*CvarRow, error) {
	var row CvarRow
	if err := s.Scan(&row.Key, &row.ValueInt, &row.ValueString, &row.MainType, &row.Title, &row.Description,
		&row.Category, &row.Ordinal); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *MySQL) ListCvars(ctx context.Context) ([]CvarRow, error) {
	rows, err := m.db.QueryContext(ctx, cvarSelectSQL()+" ORDER BY cvar_ordinal, cvar_key")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []CvarRow
	for rows.Next() {
		row, err := scanCvar(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, nil
}

func (m *MySQL) FindCvar(ctx context.Context, key string) (*CvarRow, error) {
	row := m.db.QueryRowContext(ctx, cvarSelectSQL()+" WHERE cvar_key = ?", key)
	c, err := scanCvar(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (m *MySQL) InsertCvarIfMissing(ctx context.Context, row CvarRow) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO cvars (cvar_key, cvar_value_int, cvar_value_string, cvar_main_type, cvar_title, cvar_description, cvar_category, cvar_ordinal)
		 VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   cvar_title = VALUES(cvar_title),
		   cvar_description = VALUES(cvar_description),
		   cvar_category = VALUES(cvar_category),
		   cvar_ordinal = VALUES(cvar_ordinal)`,
		row.Key, row.ValueInt, row.ValueString, row.MainType, row.Title, row.Description, row.Category, row.Ordinal)
	return err
}

func (m *MySQL) UpdateCvar(ctx context.Context, key string, valueInt int, valueString string) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE cvars SET cvar_value_int = ?, cvar_value_string = NULLIF(?, '') WHERE cvar_key = ? LIMIT 1`,
		valueInt, valueString, key)
	return err
}
