package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
    "github.com/gofiber/fiber/v3/middleware/logger"
    "github.com/gofiber/fiber/v3/middleware/recover"
    "github.com/gofiber/websocket/v2"
    "github.com/joho/godotenv"
    "github.com/messenger/backend/internal/handlers"
    "github.com/messenger/backend/internal/middleware"
    "github.com/messenger/backend/pkg/auth"
    "github.com/messenger/backend/pkg/cache"
    "github.com/messenger/backend/pkg/database"
    "github.com/messenger/backend/pkg/media"
)

func main() {
    godotenv.Load()

    appEnv := getEnv("APP_ENV", "development")
    appPort := getEnv("APP_PORT", "8080")
    logLevel := getEnv("LOG_LEVEL", "info")

    log.Printf("Starting Messenger API Server (Environment: %s)", appEnv)

    db, err := database.Initialize()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    log.Println("✓ Database connected")

    redisClient, err := cache.Initialize()
    if err != nil {
        log.Fatalf("Failed to initialize Redis: %v", err)
    }
    log.Println("✓ Redis connected")

    jwtSecret := getEnv("JWT_SECRET", "changeme_jwt_secret_key")
    if jwtSecret == "changeme_jwt_secret_key" && appEnv == "production" {
        log.Fatal("JWT_SECRET must be set in production")
    }
    auth.Initialize(jwtSecret)
    log.Println("✓ JWT initialized")

    uploader := media.NewMediaUploader()
    if err := uploader.Initialize(); err != nil {
        log.Printf("Warning: Failed to initialize media uploader: %v", err)
    } else {
        log.Println("✓ Media upload directory initialized")
    }

    rateLimiter := middleware.NewRateLimiter(redisClient)
    lastSeenMiddleware := middleware.NewLastSeenMiddleware(db, redisClient)

    app := fiber.New(fiber.Config{
        AppName:               "Messenger API v1.0.0",
        ServerHeader:          "",
        DisableStartupMessage: false,
        ReadTimeout:           time.Second * 30,
        WriteTimeout:          time.Second * 30,
        IdleTimeout:           time.Second * 60,
        BodyLimit:             100 * 1024 * 1024,
        EnablePrintRoutes:     appEnv == "development",
        ErrorHandler:          middleware.ErrorHandler(),
    })

    app.Use(recover.New())
    app.Use(logger.New(logger.Config{
        Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
        Output: os.Stdout,
    }))

    corsOrigins := getEnv("CORS_ORIGINS", "*")
    app.Use(cors.New(cors.Config{
        AllowOrigins:     corsOrigins,
        AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
        MaxAge:           3600,
    }))

    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":    "ok",
            "timestamp": time.Now().Unix(),
            "service":   "messenger-api",
        })
    })

    app.Static("/uploads", "./uploads")

    api := app.Group("/api/v1")

    authHandler := handlers.NewAuthHandler(db, redisClient)
    api.Post("/auth/register", authHandler.Register)
    api.Post("/auth/login", rateLimiter.LoginRateLimit(), authHandler.Login)
    api.Post("/auth/refresh", authHandler.RefreshToken)
    api.Post("/auth/logout", auth.Protected(), authHandler.Logout)

    userHandler := handlers.NewUserHandler(db, redisClient)

    users := api.Group("/users")
    users.Get("/me", auth.Protected(), lastSeenMiddleware.UpdateLastSeen(), userHandler.GetMe)
    users.Get("/:user_id", auth.Protected(), lastSeenMiddleware.UpdateLastSeen(), userHandler.GetUser)
    users.Patch("/me", auth.Protected(), lastSeenMiddleware.UpdateLastSeen(), userHandler.UpdateProfile)
    users.Patch("/me/password", auth.Protected(), lastSeenMiddleware.UpdateLastSeen(), userHandler.ChangePassword)
    users.Delete("/me", auth.Protected(), userHandler.DeleteAccount)

    messageHandler := handlers.NewMessageHandler(db, redisClient)
    messages := api.Group("/messages", auth.Protected(), lastSeenMiddleware.UpdateLastSeen())
    messages.Post("/", messageHandler.SendMessage)
    messages.Post("/upload", rateLimiter.UploadRateLimit(), messageHandler.SendMediaMessage)
    messages.Get("/:id", messageHandler.GetMessage)
    messages.Patch("/:id", messageHandler.EditMessage)
    messages.Delete("/:id", messageHandler.DeleteMessage)

    api.Get("/media/*", auth.Protected(), messageHandler.GetMediaFile)

    chatHandler := handlers.NewChatHandler(db, redisClient)
    chats := api.Group("/chats", auth.Protected(), lastSeenMiddleware.UpdateLastSeen())
    chats.Post("/", chatHandler.CreateChat)
    chats.Get("/", chatHandler.GetUserChats)
    chats.Get("/dm/:user_id", chatHandler.GetOrCreateDM)
    chats.Get("/:id", chatHandler.GetChat)
    chats.Get("/:id/messages", chatHandler.GetChatMessages)
    chats.Post("/:id/read", chatHandler.MarkAsRead)
    chats.Post("/:id/members", chatHandler.AddMember)
    chats.Delete("/:id/members/:userId", chatHandler.RemoveMember)

    channelHandler := handlers.NewChannelHandler(db)
    channels := api.Group("/channels", auth.Protected(), lastSeenMiddleware.UpdateLastSeen())
    channels.Post("/", channelHandler.CreateChannel)
    channels.Get("/:id", channelHandler.GetChannel)
    channels.Post("/:id/subscribe", channelHandler.Subscribe)
    channels.Delete("/:id/subscribe", channelHandler.Unsubscribe)

    subscriptionHandler := handlers.NewSubscriptionHandler(db)
    subscriptions := api.Group("/subscriptions", auth.Protected(), lastSeenMiddleware.UpdateLastSeen())
    subscriptions.Post("/purchase", subscriptionHandler.PurchaseSubscription)
    subscriptions.Get("/me", subscriptionHandler.GetMySubscription)

    wsHandler := handlers.NewWebSocketHandler(db, redisClient)

    callHandler := handlers.NewCallHandler(db, redisClient, wsHandler)
    calls := api.Group("/calls", auth.Protected(), lastSeenMiddleware.UpdateLastSeen())
    calls.Post("/", callHandler.InitiateCall)
    calls.Get("/ice-servers", callHandler.GetICEServers)
    calls.Get("/:call_id", callHandler.GetCall)
    calls.Patch("/:call_id", callHandler.RespondToCall)
    calls.Delete("/:call_id", callHandler.EndCall)
    calls.Post("/:call_id/signal", callHandler.SaveCallSignal)
    app.Use("/ws", func(c fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
    app.Get("/ws", websocket.New(wsHandler.HandleWebSocket))

    app.Use(middleware.NotFoundHandler())

    go func() {
        if logLevel == "debug" {
            log.Printf("Starting server on port %s", appPort)
        }
        if err := app.Listen(fmt.Sprintf(":%s", appPort)); err != nil {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down gracefully...")
    if err := app.Shutdown(); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

    if sqlDB, err := db.DB(); err == nil {
        sqlDB.Close()
    }
    redisClient.Close()

    log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
