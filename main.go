package main

import (
	"chat.service/configuration"
	"chat.service/database"
	"chat.service/integration/entity"
	"chat.service/mapper"
	"chat.service/operations/subscribe"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"time"
)

func main() {
	configuration.ShouldParseViperConfig()
	couchbaseConfig := configuration.ShouldParseCouchbaseConfig()
	cluster := database.ShouldGetCluster(couchbaseConfig)
	if err := cluster.WaitUntilReady(5*time.Second, nil); err != nil {
		panic(err)
	}

	e := echo.New()

	e.Use(middleware.StaticWithConfig(
		middleware.StaticConfig{
			Root:   "public",
			Index:  "index.html",
			Browse: false,
			HTML5:  true,
		},
	))

	// todo вынесу отдельно
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2019

	schema, err := compiler.Compile("contracts/woop-socket-message.json")
	if err != nil {
		panic(err)
	}

	socketSettings := &subscribe.SubscribeOperationSettings{
		SocketRequestMapper: mapper.JsonSocketRequestMapper(schema),
		Entity:              entity.Handlers(cluster),
	}

	e.GET("/ws", subscribe.OpenWebSocketConnection(socketSettings))

	e.Logger.Fatal(e.Start(":1323"))
}
