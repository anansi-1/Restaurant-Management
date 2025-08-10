package controllers


// incomingRoutes.GET("/users", controller.GetUsers())
// incomingRoutes.GET("/users/:user_id", controller.GetUser())
// incomingRoutes.POST("/users/signup", controller.SignUp())
// incomingRoutes.POST("/users/login", controller.Login())


import(
	"github.com/gin-gonic/gin"
)


func GetUsers() gin.HandlerFunc{

	return func(c *gin.Context) {

	}
	
}

func GetUser() gin.HandlerFunc {

	return  func(ctx *gin.Context) {

	}
}

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func Login() gin.HandlerFunc{
	return func(ctx *gin.Context) {}
}

