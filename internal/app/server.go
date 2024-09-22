package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ry461ch/loyalty_system/internal/crontasks/orders/enricher"
	"github.com/ry461ch/loyalty_system/internal/components/orders/initial"
	"github.com/ry461ch/loyalty_system/internal/config"
	"github.com/ry461ch/loyalty_system/internal/handlers/initial"
	"github.com/ry461ch/loyalty_system/internal/router"
	"github.com/ry461ch/loyalty_system/internal/services/initial"
	"github.com/ry461ch/loyalty_system/internal/storage/postgres"
	"github.com/ry461ch/loyalty_system/pkg/logging"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
)

type Server struct {
	cfg               *config.Config
	pgStorage    	  *pgstorage.PGStorage
	orderEnricher     *orderenricher.OrderEnricher
	server            *http.Server
}

func NewServer(cfg *config.Config) *Server {
	logging.Initialize(cfg.LogLevel)

	// initialize storage
	pgStorage := pgstorage.NewPGStorage(cfg.DBDsn)
	authenticator := authentication.NewAuthenticator(cfg.JWTSecretKey, cfg.TokenExp)
	services := servicesimpl.NewServices(pgStorage.BalanceStorage, pgStorage.WithdrawalStorage, pgStorage.UserStorage, pgStorage.OrderStorage, authenticator)
	handlers := handlersimpl.NewHandlers(services.MoneyService, services.OrderService, services.UserService)
	router := router.NewRouter(handlers.AuthHandlers, handlers.MoneyHandlers, handlers.OrdersHandlers, authenticator)
	orderComponents := ordercomponents.NewOrderComponents(cfg, services.OrderService)
	orderEnricher := orderenricher.NewOrderEnricher(orderComponents.Getter, orderComponents.Sender, orderComponents.Updater, cfg)

	server := &http.Server{Addr: cfg.Addr.Host + ":" + strconv.FormatInt(cfg.Addr.Port, 10), Handler: router}

	return &Server{
		cfg:           cfg,
		pgStorage: pgStorage,
		orderEnricher: orderEnricher,
		server:        server,
	}
}

func (s *Server) Run() {
	err := s.pgStorage.Init(context.Background())
	if err != nil {
		logging.Logger.Warnln("Db wasn't initialized")
	}
	defer s.pgStorage.Close()

	var wg sync.WaitGroup
	wg.Add(3)

	// run server
	go func() {
		logging.Logger.Info("Server is running: ", s.cfg.Addr.String())
		s.server.ListenAndServe()
		wg.Done()
	}()

	// run crontasks
	orderEnricherCtx, orderEnricherCtxCancel := context.WithCancel(context.Background())
	go func() {
		s.orderEnricher.Run(orderEnricherCtx)
		wg.Done()
	}()

	// wait for interrupting signal
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		s.server.Shutdown(context.Background())
		orderEnricherCtxCancel()
		wg.Done()
	}()

	wg.Wait()
}
