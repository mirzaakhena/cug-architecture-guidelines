package wiring

import (
	"net/http"
	"project/utils"

	"gorm.io/gorm"
)

func SetupWiring(mux *http.ServeMux, db *gorm.DB, apiPrinter *utils.ApiPrinter) {

	// Wire up the Todo components

	// Initialize gateways
	// someGateway1 := gateway.ImpSomeGateway1(sc1)
	// someGateway2 := gateway.ImpSomeGateway2(sc1)
	// someGateway3 := gateway.ImpSomeGateway3(sc1)

	// Initialize middleware
	// someGatewayUnderMiddleware := middleware.ImpSomeMiddleware(someGateway3)

	// Initialize usecases
	// someUseCase := usecase.ImplSomeUseCase(
	// 	someGateway1,
	// 	someGateway2,
	// 	someGatewayUnderMiddleware,
	// )

	// Initialize controllers
	// controller.SomeController(sc2, someUseCase)

}
