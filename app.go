package mvp

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/mvpjobs"
	"github.com/andreyvit/mvp/mvplive"
	"github.com/andreyvit/mvp/mvputil"
	"github.com/andreyvit/mvp/postmark"
	"github.com/uptrace/bunrouter"
)

type AppOptions struct {
	Context context.Context
	Logf    func(format string, v ...interface{})
}

type AppBehaviors struct {
	IsTesting           bool
	ServeAssetsFromDisk bool
	CrashOnPanic        bool
	PrettyJSON          bool
	DisableRateLimits   bool
	AllowInsecureHttp   bool
}

type App struct {
	ValueSet
	DBSchema      edb.Schema
	JobSchema     *mvpjobs.Schema
	Configuration *Configuration
	Settings      *Settings
	Hooks         Hooks
	BaseURL       *url.URL
	IPForwarding  *mvputil.IPForwarding

	stopApp func()
	logf    func(format string, args ...any)

	routesByName map[string]*Route
	domainRouter *DomainRouter
	siteRouters  map[*Site]*bunrouter.Router

	staticFS     fs.FS
	viewsFS      fs.FS
	templates    *template.Template
	templatesDev atomic.Value

	db                  *edb.DB
	gen                 *flake.Gen
	dbMonitoringOptions map[*edb.Table]edb.ChangeFlags

	methodsByName     map[string]*MethodImpl
	jobsByKind        map[*mvpjobs.Kind]*JobImpl
	ephemeralJobQueue EphemeralJobQueue
	liveQueue         *mvplive.Queue

	postmrk *postmark.Caller

	rateLimiters map[RateLimitPreset]map[RateLimitGranularity]*RateLimiter

	// rateLimiters map[string]
}

type AppInit struct {
	app *App
}

func (app *App) Initialize(settings *Settings, opt AppOptions) {
	if opt.Logf == nil {
		opt.Logf = log.Printf
	}
	if opt.Context == nil {
		opt.Context = context.Background()
	}
	if settings.Env == "" {
		panic("settings.Env not set")
	}
	if settings.AppID == "" {
		panic(fmt.Errorf("%s: AppID not configured", settings.Configuration.ConfigFileName))
	}
	if len(settings.JWTIssuers) == 0 {
		settings.JWTIssuers = []string{settings.AppID}
	}

	ctx, stopApp := context.WithCancel(opt.Context)
	_ = ctx

	app.ValueSet = newValueSet()
	app.Configuration = settings.Configuration
	app.Settings = settings
	app.routesByName = make(map[string]*Route)
	app.logf = opt.Logf
	app.stopApp = stopApp

	if app.BaseURL == nil && settings.BaseURL != "" {
		app.BaseURL = must(url.Parse(settings.BaseURL))
	}
	app.IPForwarding = &mvputil.IPForwarding{
		ProxyIPs: mvputil.MustParseCIDRs(settings.ForwardingProxyCIDRs),
	}

	app.initEphemeralJobs()
	app.initLive()

	app.JobSchema = &mvpjobs.Schema{}
	app.addModule(builtinModule)
	for _, mod := range app.Settings.Configuration.Modules {
		app.addModule(mod)
	}
	for _, kind := range app.JobSchema.Kinds() {
		if kind.IsPersistent() {
			app.JobImpl(kind)
		}
	}
	log.Printf("app jobs: %v", app.JobSchema.PersistentKindNames())

	initAppDB(app, &opt)
	initViews(app, &opt)

	app.postmrk = &postmark.Caller{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Credentials: app.Settings.Postmark,
	}

	initRateLimiting(app)
	initRouting(app)

	init := AppInit{app}
	runHooksFwd2(app.Hooks.initApp, app, &init)

	{
		rc := NewRC(ctx, app, "init")
		defer rc.Close()
		allMigrations := collectMigrations(app, rc)

		rc.MustWrite(func() {
			runHooksFwd2(app.Hooks.initDB, app, rc)
			executeMigrations(allMigrations, rc)
		})
	}
}

func (app *App) addModule(mod *Module) {
	log.Printf("app including module %s", mod.Name)
	if mod.DBSchema != nil {
		app.DBSchema.Include(mod.DBSchema)
	}
	if mod.JobSchema != nil {
		log.Printf("... app including module %s jobs %v and %v", mod.Name, mod.JobSchema.PersistentKindNames(), mod.JobSchema.EphemeralKindNames())
		app.JobSchema.Include(mod.JobSchema)
	}
}

func (app *App) Close() {
	app.stopApp()
	runHooksRev1(app.Hooks.closeApp, app)
	closeAppDB(app)
}

func (init *AppInit) MonitorDBChanges(tbl *edb.Table, flags edb.ChangeFlags) {
	app := init.app
	if app.dbMonitoringOptions == nil {
		app.dbMonitoringOptions = make(map[*edb.Table]edb.ChangeFlags)
	}
	app.dbMonitoringOptions[tbl] |= flags | edb.ChangeFlagNotify
}
