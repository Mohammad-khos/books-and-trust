package main

import (
	"context"
	_ "go.uber.org/automaxprocs"
)

//	@title			Books & Trust API Gateway
//	@version		1.0
//	@description	این سرویس به عنوان درگاه اصلی برای دسترسی به میکروسرویس‌های امانت و کاربران عمل می‌کند.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

// @BasePath					/api/v1
// @host						localhost:8081
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				تایپ کنید: Bearer {token}
func main() {
	//create application instance
	app, cleanUp := NewApplication()
	defer cleanUp()
	//mount application
	app.mount()
	//run server
	if err := app.Run(context.Background()); err != nil {
		app.Logger.Fatalw("application stopped", "error", err)
	}
}
