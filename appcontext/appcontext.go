package appcontext

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/lithictech/go-aperitif/logctx"
	"github.com/lithictech/webhookdb-cli/client"
	"github.com/lithictech/webhookdb-cli/config"
	"github.com/lithictech/webhookdb-cli/statemachine"
	"github.com/sirupsen/logrus"
	"os"
)

type AppContext struct {
	Config       config.Config
	Resty        *resty.Client
	StateMachine statemachine.StateMachine
	logger       *logrus.Entry
}

func (ac AppContext) Logger() *logrus.Entry {
	return ac.logger
}

func New(command string, cfg config.Config) (ac AppContext, err error) {
	ac.Config = cfg
	ac.Resty = newResty(cfg)
	ac.StateMachine = client.NewStateMachine()
	if ac.logger, err = logctx.NewLogger(logctx.NewLoggerInput{
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		File:      cfg.LogFile,
		BuildSha:  config.BuildSha,
		BuildTime: config.BuildTime,
		Fields:    logrus.Fields{"command": command},
	}); err != nil {
		return
	}
	// TODO: For now, always use stderr instead of stdout because we are running this as a CLI,
	// not an application, and want the caller to be able to collect logs easily.
	// But we should make this better configurable, maybe with a change to logctx.
	ac.logger.Logger.SetOutput(os.Stderr)
	return
}

func NewTestContext() AppContext {
	cfg := config.LoadConfig()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	ac := AppContext{
		logger: logger.WithFields(nil),
		Config: cfg,
		Resty:  newResty(cfg),
	}
	return ac
}

const ctxKey = "appcontext"

func InContext(parent context.Context, ac AppContext) context.Context {
	return context.WithValue(parent, ctxKey, ac)
}

func FromContext(c context.Context) AppContext {
	return c.Value(ctxKey).(AppContext)
}

func newResty(cfg config.Config) *resty.Client {
	r := resty.New().
		SetHostURL(cfg.ApiHost).
		SetHeader(
			"User-Agent",
			fmt.Sprintf("WebhookdbCLI/%s built %s", config.BuildSha, config.BuildTime),
		)
	r.Debug = cfg.Debug
	return r
}
