package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ry461ch/loyalty_system/internal/components/orders"
	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/crontasks/orders/enricher"
	"github.com/ry461ch/loyalty_system/internal/handlers"
	"github.com/ry461ch/loyalty_system/internal/router"
	"github.com/ry461ch/loyalty_system/internal/services"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
	"github.com/ry461ch/loyalty_system/pkg/logging"
)

type Server struct {
	cfg           *config.Config
	pgStorage     *pgstorage.PGStorage
	orderEnricher *orderenricher.OrderEnricher
	server        *http.Server
}

func NewServer(cfg *config.Config) *Server {
	logging.Initialize(cfg.LogLevel)

	// initialize storage
	pgStorage := pgstorage.NewPGStorage(cfg.DBDsn)
	authenticator := authentication.NewAuthenticator(cfg.JWTSecretKey, cfg.TokenExp)
	services := servicesimpl.NewServices(pgStorage.BalanceStorage, pgStorage.WithdrawalStorage, pgStorage.UserStorage, pgStorage.OrderStorage, authenticator)
	handlers := handlersimpl.NewHandlers(services.MoneyService, services.OrderService, services.UserService)
	router := router.NewRouter(handlers.AuthHandlers, handlers.MoneyHandlers, handlers.OrdersHandlers, authenticator)
	orderComponents := ordercomponentsimpl.NewOrderComponents(cfg, services.OrderService)
	orderEnricher := orderenricher.NewOrderEnricher(orderComponents.Getter, orderComponents.Sender, orderComponents.Updater, cfg)

	server := &http.Server{Addr: cfg.Addr.Host + ":" + strconv.FormatInt(cfg.Addr.Port, 10), Handler: router}

	return &Server{
		cfg:           cfg,
		pgStorage:     pgStorage,
		orderEnricher: orderEnricher,
		server:        server,
	}
}

func (s *Server) Run() {
	err := s.pgStorage.Init(context.Background())
	if err != nil {
		logging.Logger.Errorf("Db wasn't initialized: %s", err.Error())
		return
	}
	defer s.pgStorage.Close()
	logging.Logger.Infof("Server: intiated db")

	var wg sync.WaitGroup
	wg.Add(3)

	// run server
	go func() {
		logging.Logger.Info("Server is running: ", s.cfg.Addr.String())
		err = s.server.ListenAndServe()
		if err != nil {
			logging.Logger.Errorf("Server: something went wrong while serving: %v", err)
		}
		logging.Logger.Infof("Server: stopped")
		wg.Done()
	}()

	// run crontasks
	orderEnricherCtx, orderEnricherCtxCancel := context.WithCancel(context.Background())
	go func() {
		logging.Logger.Infof("Server: order enricher started")
		err = s.orderEnricher.Run(orderEnricherCtx)
		if err != nil {
			logging.Logger.Errorf("Server: something went wrong while running order enricher: %v", err)
		}
		logging.Logger.Infof("Server: order enricher stopped")
		wg.Done()
	}()

	// wait for interrupting signal
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		logging.Logger.Infof("Server: got interrupt signal")
		err = s.server.Shutdown(context.Background())
		if err != nil {
			logging.Logger.Errorf("Server: something went wrong while shutting down server: %v", err)
		}
		orderEnricherCtxCancel()
		wg.Done()
	}()

	wg.Wait()
	logging.Logger.Infof("Server: done")
}
