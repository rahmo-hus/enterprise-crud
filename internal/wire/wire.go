//go:build wireinject
// +build wireinject

package wire

import (
	"enterprise-crud/internal/app"
	"enterprise-crud/internal/config"
	"enterprise-crud/internal/domain/user"
	"enterprise-crud/internal/infrastructure/database"
	"enterprise-crud/internal/presentation/http"
	"github.com/google/wire"
)

// Provider sets for dependency injection
var DatabaseProviderSet = wire.NewSet(
	database.NewConnection,
	database.NewUserRepository,
)

var ServiceProviderSet = wire.NewSet(
	user.NewUserService,
)

var HandlerProviderSet = wire.NewSet(
	http.NewUserHandler,
)

var AppProviderSet = wire.NewSet(
	app.New,
)

// InitializeApp creates a fully configured application
func InitializeApp(cfg *config.Config) (*app.App, error) {
	wire.Build(
		DatabaseProviderSet,
		ServiceProviderSet,
		HandlerProviderSet,
		AppProviderSet,
	)
	return nil, nil
}