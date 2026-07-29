package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	auth "github.com/jamesread/httpauthshim"
	"github.com/jamesread/httpauthshim/authpublic"
	"github.com/jamesread/httpauthshim/sessions"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"faridoon/service/buildinfo"
	"faridoon/service/gen/faridoon/v1/faridoonv1connect"
	"faridoon/service/internal/config"
	"faridoon/service/internal/quote"
	"faridoon/service/internal/store"
	"faridoon/service/internal/webhook"
)

type contextKey string

const httpRequestKey contextKey = "httpRequest"

func main() {
	configDir := flag.String("configdir", "", "directory containing config.yaml")
	flag.Parse()
	if *configDir != "" {
		config.SetConfigDir(*configDir)
	}

	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	cfg := config.LoadConfig()
	logConfigLoaded()

	db, err := openDB(cfg)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer func() { _ = db.Close() }()

	st := store.NewMySQL(db)
	if migErr := assertMigration(context.Background(), st, cfg.RequiredMigration); migErr != nil {
		return migErr
	}
	if cvarErr := ensureDefaultCvars(context.Background(), st, cfg.SiteTitle); cvarErr != nil {
		return fmt.Errorf("cvars: %w", cvarErr)
	}

	authCtx, err := setupAuth(cfg)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	if authCtx != nil {
		defer func() { _ = authCtx.Shutdown() }()
	}

	srv := newFaridoonServer(cfg, st, authCtx)
	mux := buildMux(srv)

	addr := config.ListenAddr(cfg)
	logrus.Infof("Starting Faridoon %s on %s", buildinfo.Version, addr)
	return serveHTTP(addr, mux)
}

func logConfigLoaded() {
	if p := config.GetConfigPath(); p != "" {
		logrus.Infof("Config loaded from %s", p)
	}
}

func newFaridoonServer(cfg *config.Config, st store.Store, authCtx *auth.AuthShimContext) *FaridoonServer {
	return &FaridoonServer{
		cfg:       cfg,
		store:     st,
		auth:      authCtx,
		formatter: quote.NewFormatter(),
		webhooks:  &webhook.Dispatcher{Store: st, Client: &http.Client{Timeout: 2 * time.Second}},
	}
}

func buildMux(srv *FaridoonServer) *http.ServeMux {
	mux := http.NewServeMux()
	path, handler := faridoonv1connect.NewFaridoonServiceHandler(srv)
	mux.Handle(path, withRequestContext(handler))
	mux.Handle("/metrics", promhttp.Handler())

	if staticDir := getStaticDir(); staticDir != "" {
		mux.Handle("/", spaFileServer(staticDir))
		logrus.Infof("Serving SPA from %s", staticDir)
	}
	return mux
}

func serveHTTP(addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(handler, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	return srv.ListenAndServe()
}

func openDB(cfg *config.Config) (*sql.DB, error) {
	port := cfg.Database.Port
	if port == 0 {
		port = 3306
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, port, cfg.Database.Name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if pingErr := db.PingContext(context.Background()); pingErr != nil {
		_ = db.Close()
		return nil, pingErr
	}
	return db, nil
}

func assertMigration(ctx context.Context, st store.Store, required string) error {
	latest, err := st.LatestMigration(ctx)
	if err != nil {
		return fmt.Errorf("connected to the database, but the migrations table could not be queried. run sql-migrate up: %w", err)
	}
	display := latest
	if display == "" {
		display = "null"
	}
	if latest != required {
		return fmt.Errorf("requires database version %s but the database is at version %s; run database migrations", required, display)
	}
	return nil
}

func setupAuth(cfg *config.Config) (*auth.AuthShimContext, error) {
	authCfg := cfg.Auth
	if authCfg == nil {
		authCfg = &authpublic.Config{}
		authCfg.LocalUsers.Enabled = true
	}
	sessionStorage := sessions.NewSessionStorage(sessions.NewYAMLPersistence())
	authCtx, err := auth.NewAuthShimContext(authCfg, sessionStorage)
	if err != nil {
		return nil, err
	}
	authCtx.AddProvider(checkUserFromFaridoonSession)
	logrus.Info("Authentication enabled (httpauthshim, DB-backed sessions)")
	return authCtx, nil
}

func checkUserFromFaridoonSession(ac *authpublic.AuthCheckingContext) *authpublic.AuthenticatedUser {
	u := &authpublic.AuthenticatedUser{}
	sid, ok := faridoonSessionCookie(ac)
	if !ok {
		return u
	}
	sess := lookupFaridoonSession(ac, sid)
	if sess == nil {
		return u
	}
	return authenticatedFromSession(u, sid, sess)
}

func faridoonSessionCookie(ac *authpublic.AuthCheckingContext) (string, bool) {
	if ac.Request == nil || ac.Config == nil {
		return "", false
	}
	cookieName := ac.Config.GetLocalSessionCookieName()
	c, err := ac.Request.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

func authenticatedFromSession(u *authpublic.AuthenticatedUser, sid string, sess *sessions.UserSession) *authpublic.AuthenticatedUser {
	u.Username = sess.Username
	u.UsergroupLine = sess.Usergroup
	u.Provider = "faridoon"
	u.SID = sid
	return u
}

func lookupFaridoonSession(ac *authpublic.AuthCheckingContext, sid string) *sessions.UserSession {
	if ac.Sessions != nil {
		return ac.Sessions.GetSession("faridoon", sid)
	}
	return sessions.GetUserSession("faridoon", sid)
}

func withRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), httpRequestKey, r)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getStaticDir() string {
	dir := os.Getenv("FARIDOON_STATIC_DIR")
	if dir == "" {
		dir = findStaticDirCandidate()
	}
	if dir == "" {
		return ""
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	return abs
}

func findStaticDirCandidate() string {
	for _, c := range []string{"../frontend/dist", "./frontend/dist", "/app/frontend"} {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	return ""
}

func spaFileServer(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	indexPath := filepath.Join(root, "index.html")
	rootAbs, _ := filepath.Abs(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		target := spaTarget(root, rootAbs, r.URL.Path)
		if target == indexPath {
			http.ServeFile(w, r, indexPath)
			return
		}
		fs.ServeHTTP(w, r)
	})
}

func spaTarget(root, rootAbs, urlPath string) string {
	indexPath := filepath.Join(root, "index.html")
	path := filepath.Clean(urlPath)
	if isRootPath(path) {
		return indexPath
	}
	fpath := filepath.Join(root, path)
	if !isSafeRegularFile(fpath, rootAbs) {
		return indexPath
	}
	return fpath
}

func isRootPath(path string) bool {
	return path == "/" || path == "."
}

func isSafeRegularFile(fpath, rootAbs string) bool {
	abs, err := filepath.Abs(fpath)
	if err != nil {
		return false
	}
	if !strings.HasPrefix(abs, rootAbs) {
		return false
	}
	fi, err := os.Stat(fpath)
	return err == nil && fi.Mode().IsRegular()
}
