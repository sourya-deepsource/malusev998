package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/malusev998/malusev998/server/container"
	"github.com/malusev998/malusev998/server/handlers"
	"github.com/malusev998/malusev998/server/utils"
)

func RegisterRouter(c *container.Container, app *fiber.App) {
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return utils.UniqueStringGenerator(64)
		},
		ContextKey: "request_id",
	}), recover.New(), compress.New(compress.Config{
		Next:  nil,
		Level: compress.LevelBestSpeed,
	}))

	app.Use(pprof.New())

	app.Static("/public", "./public", fiber.Static{
		Browse:    false,
		Compress:  true,
		ByteRange: true,
	})

	globalGroup := app.Group("")

	globalGroup.Use(csrf.New(csrf.Config{
		KeyLookup:      "cookie:csrf_cookie",
		ContextKey:     "csrf_token",
		Storage:        c.GetStorage(0),
		CookieName:     "csrf_cookie",
		CookieDomain:   c.Config.Csrf.CookieDomain,
		CookieSecure:   c.Config.Csrf.Secure,
		CookieHTTPOnly: true,
		KeyGenerator:   utils.DefaultStringGenerator,
	}))

	globalGroup.Get("/monitor", monitor.New())

	registerHomeRoutes(c, globalGroup)
	registerSubscribeRoutes(c, globalGroup)
	registerContactRoutes(c, globalGroup.Group("/contact"))

	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404) // => 404 "Not Found"
	})
}

func registerHomeRoutes(c *container.Container, app fiber.Router) {
	home := handlers.Home{}

	app.Get("/", home.Home)
	app.Get("/about", home.About)
}

func registerSubscribeRoutes(c *container.Container, app fiber.Router) {
	sub := handlers.Subscribe{
		Service: c.GetSubscriptionService(),
	}

	limitMiddleware := limiter.New(limiter.Config{
		Max:          1,
		Storage:      c.GetStorage(0),
		KeyGenerator: utils.LimiterKeyGenerator(c.Config.Key),
	})

	app.Post("/subscribe", limitMiddleware, sub.Subscribe)
	app.Get("/unsubscribe", sub.Unsubscribe)
}

func registerContactRoutes(c *container.Container, router fiber.Router) {
	contact := handlers.Contact{
		Service: c.GetContactService(),
	}

	router.Get("/", contact.Index)
	router.Post("/", contact.Message)
}
