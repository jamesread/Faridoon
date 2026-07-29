package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// MySQL reserved identifiers used as column names in Faridoon's schema.
const (
	colGroup = "`group`"
	colKey   = "`key`"
)

type QuoteRow struct {
	Content            string
	Created            string
	SyntaxHighlighting string
	ID                 int
	VoteCount          int
	Approval           int
	Approved           bool
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

type WebhookRow struct {
	URL     string
	Event   string
	Secret  string
	Created string
	Updated string
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

type Store interface {
	LatestMigration(ctx context.Context) (string, error)
	FindQuote(ctx context.Context, id int) (*QuoteRow, error)
	FindQuoteRaw(ctx context.Context, id int) (*QuoteRow, error)
	ListApproved(ctx context.Context, order string, page, perPage int) ([]QuoteRow, int, error)
	ListPending(ctx context.Context) ([]QuoteRow, error)
	CountPending(ctx context.Context) (int, error)
	CreateQuote(ctx context.Context, content string, approval int, syntax string) (int, error)
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
	GroupsWithPermissions(ctx context.Context) ([]GroupRow, map[int][]GroupPermissionRow, error)
	CreateGroup(ctx context.Context, title string) (int, error)
	DeleteGroup(ctx context.Context, id int) error
	ListPermissions(ctx context.Context) ([]PermissionRow, error)
	GrantPermission(ctx context.Context, groupID, permissionID int) error
	RevokePermission(ctx context.Context, groupID, permissionID int) error
	ListWebhooks(ctx context.Context) ([]WebhookRow, error)
	FindWebhook(ctx context.Context, id int) (*WebhookRow, error)
	EnabledWebhooksForEvent(ctx context.Context, event string) ([]WebhookRow, error)
	CreateWebhook(ctx context.Context, url, secret, event string, enabled bool) (int, error)
	UpdateWebhook(ctx context.Context, id int, fields map[string]any) error
	DeleteWebhook(ctx context.Context, id int) error
	InsertLog(ctx context.Context, entry LogEntry) error
	ListLogs(ctx context.Context, page, pageSize int) ([]LogEntry, int, error)
}

type MySQL struct {
	db *sql.DB
}

func NewMySQL(db *sql.DB) *MySQL {
	return &MySQL{db: db}
}

func maxMigrationID(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	latest := ids[0]
	for _, id := range ids[1:] {
		if id > latest {
			latest = id
		}
	}
	return latest
}

func scanMigrationIDs(rows *sql.Rows) ([]string, error) {
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MySQL) LatestMigration(ctx context.Context) (string, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT id FROM migrations")
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	ids, err := scanMigrationIDs(rows)
	if err != nil {
		return "", err
	}
	return maxMigrationID(ids), nil
}

func quoteSelectSQL() string {
	return `
SELECT q.id, q.content, COALESCE(DATE_FORMAT(q.created, '%Y-%m-%d'), ''),
  q.approval, COALESCE(q.syntaxHighlighting, ''), COALESCE(v.voteCount, 0)
FROM quotes q
LEFT JOIN (
  SELECT quote, COALESCE(SUM(delta), 0) AS voteCount FROM votes GROUP BY quote
) v ON v.quote = q.id`
}

func scanQuote(s interface{ Scan(...any) error }) (*QuoteRow, error) {
	var q QuoteRow
	var approval int
	if err := s.Scan(&q.ID, &q.Content, &q.Created, &approval, &q.SyntaxHighlighting, &q.VoteCount); err != nil {
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
		`SELECT id, content, COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), ''), approval, COALESCE(syntaxHighlighting, ''), 0 FROM quotes WHERE id = ?`, id)
	q, err := scanQuote(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return q, err
}

func (m *MySQL) ListApproved(ctx context.Context, order string, page, perPage int) ([]QuoteRow, int, error) {
	var total int
	if err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quotes WHERE approval = 1`).Scan(&total); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	rows, err := m.db.QueryContext(ctx, quoteSelectSQL()+" WHERE q.approval = 1 "+approvedOrderSQL(order)+" LIMIT ? OFFSET ?", perPage, offset)
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

func (m *MySQL) CreateQuote(ctx context.Context, content string, approval int, syntax string) (int, error) {
	res, err := m.db.ExecContext(ctx,
		`INSERT INTO quotes (content, approval, syntaxHighlighting, created) VALUES (?, ?, NULLIF(?, ''), NOW())`,
		content, approval, syntax)
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
	if _, err := tx.ExecContext(ctx, `DELETE FROM votes WHERE quote = ? AND user = ?`, quoteID, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO votes (quote, user, delta) VALUES (?, ?, ?)`, quoteID, userID, delta); err != nil {
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
	return "SELECT u.id, u.username, u.password, COALESCE(u." + colGroup + ", 0), COALESCE(g.title, '') FROM users u LEFT JOIN groups g ON u." + colGroup + " = g.id"
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
		"INSERT INTO users (username, password, "+colGroup+", registered) VALUES (?, ?, ?, NOW())",
		username, passwordHash, groupID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) UpdatePassword(ctx context.Context, userID int, passwordHash string) error {
	_, err := m.db.ExecContext(ctx, `UPDATE users SET password = ? WHERE id = ?`, passwordHash, userID)
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

func (m *MySQL) loadPrivKeys(ctx context.Context, query string, arg any) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, query, arg)
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

func mergePrivKeys(into map[string]struct{}, keys []string) {
	for _, k := range keys {
		into[k] = struct{}{}
	}
}

func privKeysToSlice(privs map[string]struct{}) []string {
	out := make([]string, 0, len(privs))
	for k := range privs {
		out = append(out, k)
	}
	return out
}

func (m *MySQL) UserPrivileges(ctx context.Context, userID, groupID int) ([]string, error) {
	privs := map[string]struct{}{}
	q1 := "SELECT p." + colKey + " FROM privileges_u pu JOIN permissions p ON p.id = pu.permission WHERE pu.user = ?"
	userKeys, err := m.loadPrivKeys(ctx, q1, userID)
	if err != nil {
		return nil, err
	}
	mergePrivKeys(privs, userKeys)
	if groupID > 0 {
		q2 := "SELECT p." + colKey + " FROM privileges_g pg JOIN permissions p ON p.id = pg.permission WHERE pg." + colGroup + " = ?"
		groupKeys, groupErr := m.loadPrivKeys(ctx, q2, groupID)
		if groupErr != nil {
			return nil, groupErr
		}
		mergePrivKeys(privs, groupKeys)
	}
	return privKeysToSlice(privs), nil
}

func (m *MySQL) ListGroups(ctx context.Context) ([]GroupRow, error) {
	rows, err := m.db.QueryContext(ctx, `SELECT id, title FROM groups ORDER BY id`)
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
	perms := map[int][]GroupPermissionRow{}
	for _, g := range groups {
		gp, loadErr := m.loadGroupPerms(ctx, g.ID)
		if loadErr != nil {
			return nil, nil, loadErr
		}
		perms[g.ID] = gp
	}
	return groups, perms, nil
}

func (m *MySQL) CreateGroup(ctx context.Context, title string) (int, error) {
	res, err := m.db.ExecContext(ctx, `INSERT INTO groups (title) VALUES (?)`, title)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (m *MySQL) DeleteGroup(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM groups WHERE id = ? LIMIT 1`, id)
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

func (m *MySQL) ListWebhooks(ctx context.Context) ([]WebhookRow, error) {
	rows, err := m.db.QueryContext(ctx, webhookSelectSQL()+" ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWebhooks(rows)
}

func webhookSelectSQL() string {
	return `SELECT id, url, event, secret, enabled,
		 COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), ''),
		 COALESCE(DATE_FORMAT(updated, '%Y-%m-%d %H:%i:%s'), '')
		 FROM webhooks`
}

func (m *MySQL) FindWebhook(ctx context.Context, id int) (*WebhookRow, error) {
	row := m.db.QueryRowContext(ctx, webhookSelectSQL()+" WHERE id = ?", id)
	w, err := scanWebhook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return w, err
}

func (m *MySQL) EnabledWebhooksForEvent(ctx context.Context, event string) ([]WebhookRow, error) {
	rows, err := m.db.QueryContext(ctx, webhookSelectSQL()+" WHERE event = ? AND enabled = 1", event)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanWebhooks(rows)
}

func (m *MySQL) CreateWebhook(ctx context.Context, url, secret, event string, enabled bool) (int, error) {
	en := 0
	if enabled {
		en = 1
	}
	now := time.Now()
	res, err := m.db.ExecContext(ctx,
		`INSERT INTO webhooks (url, secret, event, enabled, created, updated) VALUES (?, ?, ?, ?, ?, ?)`,
		url, secret, event, en, now, now)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func applyWebhookStringFields(cur *WebhookRow, fields map[string]any) {
	if u, ok := fields["url"].(string); ok {
		cur.URL = u
	}
	if s, ok := fields["secret"].(string); ok {
		cur.Secret = s
	}
	if e, ok := fields["event"].(string); ok {
		cur.Event = e
	}
}

func applyWebhookEnabled(cur *WebhookRow, fields map[string]any) {
	if en, ok := fields["enabled"].(int); ok {
		cur.Enabled = en == 1
		return
	}
	if en, ok := fields["enabled"].(bool); ok {
		cur.Enabled = en
	}
}

func (m *MySQL) UpdateWebhook(ctx context.Context, id int, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	cur, err := m.FindWebhook(ctx, id)
	if err != nil {
		return err
	}
	if cur == nil {
		return nil
	}
	applyWebhookStringFields(cur, fields)
	applyWebhookEnabled(cur, fields)
	en := 0
	if cur.Enabled {
		en = 1
	}
	_, err = m.db.ExecContext(ctx,
		`UPDATE webhooks SET url = ?, secret = ?, event = ?, enabled = ?, updated = ? WHERE id = ? LIMIT 1`,
		cur.URL, cur.Secret, cur.Event, en, time.Now(), id)
	return err
}

func (m *MySQL) DeleteWebhook(ctx context.Context, id int) error {
	_, err := m.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id = ? LIMIT 1`, id)
	return err
}

func scanWebhook(s interface{ Scan(...any) error }) (*WebhookRow, error) {
	var w WebhookRow
	var enabled int
	if err := s.Scan(&w.ID, &w.URL, &w.Event, &w.Secret, &enabled, &w.Created, &w.Updated); err != nil {
		return nil, err
	}
	w.Enabled = enabled == 1
	return &w, nil
}

func scanWebhooks(rows *sql.Rows) ([]WebhookRow, error) {
	var out []WebhookRow
	for rows.Next() {
		w, err := scanWebhook(rows)
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
		`INSERT INTO logs (created, actor_user_id, actor_username, action, entity_type, entity_id, detail, ip)
		 VALUES (?, ?, NULLIF(?, ''), ?, NULLIF(?, ''), ?, NULLIF(?, ''), NULLIF(?, ''))`,
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
	if err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM logs`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, COALESCE(DATE_FORMAT(created, '%Y-%m-%d %H:%i:%s'), ''),
		 COALESCE(actor_user_id, 0), COALESCE(actor_username, ''), action,
		 COALESCE(entity_type, ''), COALESCE(entity_id, 0), COALESCE(detail, ''), COALESCE(ip, '')
		 FROM logs ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset)
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
