package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	logurhandler "emperror.dev/handler/logur"
	"github.com/goph/logur"
	"github.com/goph/logur/integrations/zaplog"
	"github.com/lestrrat-go/backoff"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"
	"go.uber.org/cadence/.gen/go/shared"

	"github.com/sagikazarmark/cadence-bootstrap/gen/meta"
	"github.com/sagikazarmark/cadence-bootstrap/internal/platform/backoffpolicy"
	"github.com/sagikazarmark/cadence-bootstrap/internal/platform/cadence"
	"github.com/sagikazarmark/cadence-bootstrap/internal/platform/log"
)

// Provisioned by ldflags
// nolint: gochecknoglobals
var (
	version    string
	commitHash string
	buildDate  string
)

func main() {
	v, p := viper.New(), pflag.NewFlagSet(friendlyAppName, pflag.ExitOnError)

	configure(v, p)

	p.String("config", "", "Configuration file")
	p.Bool("version", false, "Show version information")

	_ = p.Parse(os.Args[1:])

	if v, _ := p.GetBool("version"); v {
		fmt.Printf("%s version %s (%s) built on %s\n", friendlyAppName, version, commitHash, buildDate)

		os.Exit(0)
	}

	if c, _ := p.GetString("config"); c != "" {
		v.SetConfigFile(c)
	}

	err := v.ReadInConfig()
	_, configFileNotFound := err.(viper.ConfigFileNotFoundError)
	if !configFileNotFound {
		emperror.Panic(errors.Wrap(err, "failed to read configuration"))
	}

	var config configuration
	err = v.Unmarshal(&config)
	emperror.Panic(errors.Wrap(err, "failed to unmarshal configuration"))

	// Create logger (first thing after configuration loading)
	logger := log.NewLogger(config.Log)

	log.SetStandardLogger(logger)

	if configFileNotFound {
		logger.Warn("configuration file not found")
	}

	err = config.Validate()
	if err != nil {
		logger.Error(err.Error())

		os.Exit(3)
	}

	// configure error handler
	errorHandler := logurhandler.New(logger)
	defer emperror.HandleRecover(errorHandler)

	ch, err := tchannel.NewChannel("cadence-bootstrap", nil)
	emperror.Panic(err)

	ch.Peers().Add(config.Cadence.Peer())
	thriftClient := thrift.NewClient(ch, "cadence-frontend", nil)
	metaClient := meta.NewTChanMetaClient(thriftClient)

	backoffPolicy := backoffpolicy.NewPolicy(config.Backoff)

	b, cancel := backoffPolicy.Start(context.Background())
	defer cancel()

	var up bool

	for backoff.Continue(b) {
		const timeout = time.Second
		ctx, cancel := thrift.NewContext(timeout)

		val, err := metaClient.Health(ctx)
		if err != nil {
			logger.Info("Cadence is not available yet")
			logger.Debug(err.Error())

			cancel()
			continue
		}
		if !val.Ok {
			logger.Info("Cadence is not healty yet")

			cancel()
			continue
		}

		up = true
		break
	}

	if !up {
		logger.Error("Cadence hasn't become available")

		os.Exit(1)
	}

	ch.Close()

	logger.Info("Cadence is up and running")

	serviceClient, err := newServiceClient("cadence-bootstrap", config, zaplog.New(logger))
	emperror.Panic(err)

	cadenceDomains := config.Cadence.Domains

	if len(cadenceDomains) == 0 && os.Getenv("CADENCE_DOMAIN") != "" {
		v := viper.New()

		_ = v.BindEnv("name", "CADENCE_DOMAIN")
		_ = v.BindEnv("description", "CADENCE_DOMAIN_DESCRIPTION")
		_ = v.BindEnv("workflowExecutionRetentionPeriodInDays", "CADENCE_DOMAIN_RETENTION")
		_ = v.BindEnv("emitMetric", "CADENCE_DOMAIN_EMIT_METRIC")

		var domain cadence.Domain

		err := v.Unmarshal(&domain)
		emperror.Panic(err)
		emperror.Panic(domain.Validate())

		cadenceDomains = append(cadenceDomains, domain)
	}

	for _, domain := range cadenceDomains {
		logger := logur.WithFields(logger, map[string]interface{}{"domain": domain.Name})
		logger.Info("creating domain")
		logger.Debug(fmt.Sprintf("%+v", domain))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := serviceClient.RegisterDomain(ctx, &shared.RegisterDomainRequest{
			Name:                                   &domain.Name,
			Description:                            &domain.Description,
			WorkflowExecutionRetentionPeriodInDays: &domain.WorkflowExecutionRetentionPeriodInDays,
			EmitMetric:                             &domain.EmitMetric,
		})
		if _, ok := err.(*shared.DomainAlreadyExistsError); ok {
			logger.Info("domain already exists")
		} else if err != nil {
			logger.Error("failed to register Cadence domain")
			logger.Debug(err.Error())
		}

		logger.Info("domain successfully registered")

		cancel()
	}
}
