// Declaramos que este es el paquete principal (main), es decir, el archivo ejecutable que inicia el programa.
package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"multicatalogo-backend/config"
	"multicatalogo-backend/controllers"
	"multicatalogo-backend/models"
	"multicatalogo-backend/repository"
	"multicatalogo-backend/routes"
)

// func main es el punto de entrada. Aquí se CREAN las dependencias y se INYECTAN en los controladores.
func main() {
	config.LoadEnv()

	// Seguridad: el secreto del JWT es obligatorio (mínimo 32 caracteres). APP_ENV=dev muestra el campo
	// "details" de los errores; en cualquier otro valor (o sin definir) los detalles internos no se exponen.
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET no está definido o es muy corto (mínimo 32 caracteres). Agrégalo a backend/.env")
	}
	models.ExponerDetalles = config.GetEnv("APP_ENV", "prod") == "dev"

	// 1) Infraestructura: PostgreSQL (obligatoria) y Redis (opcional: sin Redis la API funciona sin caché).
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Error de base de datos: %v", err)
	}
	defer db.Close()

	var cache repository.CacheRepository = repository.NoopCache{}
	if rdb, err := config.ConnectRedis(); err != nil {
		log.Printf("ADVERTENCIA: %v. La API seguirá funcionando SIN caché.", err)
	} else {
		defer rdb.Close()
		cache = repository.NewRedisCache(rdb)
	}

	// 2) Repositorio de datos (recibe el pool). Las dos variables permiten demostrar el ROLLBACK en clase.
	repo := repository.NewPostgresRepository(db, repository.Opciones{
		SimularFalla:     config.GetEnvBool("SIMULAR_FALLA_CHECKOUT", false),
		PausaAntesCommit: time.Duration(config.GetEnvInt("CHECKOUT_PAUSA_SEGUNDOS", 0)) * time.Second,
	})

	// 3) Controladores: reciben el repositorio y la caché por interfaz (Inyección de Dependencias).
	checkout := controllers.NewCheckoutController(repo, cache)
	deps := routes.Dependencias{
		Auth:     controllers.NewAuthController(repo, cache, jwtSecret),
		Producto: controllers.NewProductoController(repo, cache),
		Red:      controllers.NewRedController(repo, cache),
		Checkout: checkout,
		Health:   controllers.NewHealthController(repo, cache),
		Pedido:   controllers.NewPedidoController(repo, cache),

		Liquidacion: controllers.NewLiquidacionController(repo),
		Imagen:      controllers.NewImagenController(repo),
		Comprobante: controllers.NewComprobanteController(repo),
		Reporte:     controllers.NewReporteController(repo),
		Usuario:     controllers.NewUsuarioController(repo),

		JWTSecret: jwtSecret,
	}

	// 4) Servidor web.
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: origenesPermitidos(),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))
	routes.SetupRoutes(app, deps)

	// Al presionar Ctrl+C (o recibir SIGTERM) apagamos Fiber con calma.
	salir := make(chan os.Signal, 1)
	signal.Notify(salir, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-salir
		log.Println("Apagando servidor...")
		_ = app.Shutdown()
	}()

	// El puerto se puede cambiar con PORT (por defecto 3000), ej.: PORT=3001 go run .
	puerto := config.GetEnv("PORT", "3000")
	if err := app.Listen(":" + puerto); err != nil {
		log.Printf("Servidor detenido: %v", err)
	}

	// Antes de cerrar la BD esperamos a que terminen las comisiones que estén calculándose en segundo plano.
	checkout.Esperar()
}

// origenesPermitidos arma la lista de orígenes CORS: los de desarrollo local más los que se indiquen en
// CORS_ORIGINS (separados por coma), p. ej. la URL del frontend desplegado en Render:
//
//	CORS_ORIGINS=https://multicatalogo-web.onrender.com
func origenesPermitidos() string {
	origenes := []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://172.17.82.108:5173"}
	for _, o := range strings.Split(config.GetEnv("CORS_ORIGINS", ""), ",") {
		// El navegador envía el origen SIN "/" final; lo quitamos por si se pegó con ella.
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			origenes = append(origenes, o)
		}
	}
	return strings.Join(origenes, ", ")
}
