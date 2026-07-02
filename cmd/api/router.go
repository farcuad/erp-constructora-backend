// cmd/api/router.go
package main

import (
	"database/sql"
	"net/http"

	"erp-constructora/internal/attendance"
	"erp-constructora/internal/budgets"
	"erp-constructora/internal/clients"
	"erp-constructora/internal/equipement"
	"erp-constructora/internal/expense"
	"erp-constructora/internal/inventory"
	"erp-constructora/internal/personnel"
	"erp-constructora/internal/project"
	"erp-constructora/internal/purchase"
	"erp-constructora/internal/suppliers"
	"erp-constructora/internal/users"
)

// SetupRoutes recibe la conexión de la BD, inicializa los módulos y devuelve el enrutador listo
func SetupRoutes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// 1. Inicializar Módulo de Usuarios y Empresas (Fase 1)
	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)

	projectRepo := project.NewRepository(db)
	projectService := project.NewService(projectRepo)
	projectHandler := project.NewHandler(projectService)

	clientRepo := clients.NewRepository(db)
	clientService := clients.NewService(clientRepo)
	clientHandler := clients.NewHandler(clientService)

	budgetRepo := budgets.NewRepository(db)
	budgetService := budgets.NewService(budgetRepo)
	budgetHandler := budgets.NewHandler(budgetService)

	expenseRepo := expense.NewRepository(db)
	expenseService := expense.NewService(expenseRepo)
	expenseHandler := expense.NewHandler(expenseService)

	purchaseRepo := purchase.NewRepository(db)
	purchaseService := purchase.NewService(purchaseRepo)
	purcharseHandler := purchase.NewHandler(purchaseService)

	supplierRepo := suppliers.NewRepository(db)
	suppilerService := suppliers.NewService(supplierRepo)
	suppliersHandler := suppliers.NewHandler(suppilerService)

	inventoryRepository := inventory.NewRepository(db)
	inventoryService := inventory.NewService(inventoryRepository)
	inventoryHandler := inventory.NewHandler(inventoryService)

	equipementRepository := equipement.NewRepository(db)
	equipementService := equipement.NewService(equipementRepository)
	equipementHandler := equipement.NewHandler(equipementService)

	personnelRespository := personnel.NewRepository(db)
	personnelService := personnel.NewService(personnelRespository)
	personnelHandler := personnel.NewHandler(personnelService)

	attendanceRepositoy := attendance.NewRepository(db)
	attendanceService := attendance.NewService(attendanceRepositoy)
	attendanceHandler := attendance.NewHandler(attendanceService)
	// Definir las rutas de este módulo
	mux.HandleFunc("POST /register", userHandler.RegisterCompanyAndAdmin)
	mux.HandleFunc("POST /login", userHandler.Login)

	mux.Handle("POST /projects", users.AuthMiddleware(http.HandlerFunc(projectHandler.Create)))
	mux.Handle("GET /projects", users.AuthMiddleware(http.HandlerFunc(projectHandler.GetAll)))

	mux.Handle("POST /clients", users.AuthMiddleware(http.HandlerFunc(clientHandler.Create)))
	mux.Handle("GET /clients", users.AuthMiddleware(http.HandlerFunc(clientHandler.GetAll)))

	mux.Handle("POST /budgets", users.AuthMiddleware(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("GET /budgets/{project_id}", users.AuthMiddleware(http.HandlerFunc(budgetHandler.GetBudgetsByProjectID)))

	mux.Handle("POST /expenses", users.AuthMiddleware(http.HandlerFunc(expenseHandler.Create)))
	mux.Handle("GET /expenses/{project_id}", users.AuthMiddleware(http.HandlerFunc(expenseHandler.GetByProject)))

	mux.Handle("POST /purcharse", users.AuthMiddleware(http.HandlerFunc(purcharseHandler.CreatePurchaseOrder)))
	mux.Handle("GET /purcharse/{project_id}", users.AuthMiddleware(http.HandlerFunc(purcharseHandler.GetOrdersByProject)))

	mux.Handle("POST /supplier", users.AuthMiddleware(http.HandlerFunc(suppliersHandler.CreateSupplier)))
	mux.Handle("GET /supplier/{project_id}", users.AuthMiddleware(http.HandlerFunc(suppliersHandler.GetAllSuppliers)))

	mux.Handle("POST /materials", users.AuthMiddleware(http.HandlerFunc(inventoryHandler.CreateMaterial)))
	mux.Handle("POST /warehouses", users.AuthMiddleware(http.HandlerFunc(inventoryHandler.CreateWarehouse)))
	mux.Handle("POST /inventory/movements", users.AuthMiddleware(http.HandlerFunc(inventoryHandler.PostMovement)))
	mux.Handle("GET /inventory/stock/{warehouse_id}", users.AuthMiddleware(http.HandlerFunc(inventoryHandler.GetStock)))

	mux.Handle("POST /equipment", users.AuthMiddleware(http.HandlerFunc(equipementHandler.CreateEquipment)))
	mux.Handle("GET /equipment", users.AuthMiddleware(http.HandlerFunc(equipementHandler.GetAll)))
	mux.Handle("POST /equipment/assignments", users.AuthMiddleware(http.HandlerFunc(equipementHandler.Assign)))
	mux.Handle("POST /equipment/maintenances", users.AuthMiddleware(http.HandlerFunc(equipementHandler.Maintenance)))

	mux.Handle("POST /positions", users.AuthMiddleware(http.HandlerFunc(personnelHandler.CreatePosition)))
	mux.Handle("POST /employees", users.AuthMiddleware(http.HandlerFunc(personnelHandler.CreateEmployee)))
	mux.Handle("GET /employees", users.AuthMiddleware(http.HandlerFunc(personnelHandler.GetEmployees)))
	mux.Handle("POST /contracts", users.AuthMiddleware(http.HandlerFunc(personnelHandler.CreateContract)))

	mux.Handle("POST /attendance", users.AuthMiddleware(http.HandlerFunc(attendanceHandler.SaveAttendance)))
	mux.Handle("GET /attendance/{project_id}", users.AuthMiddleware(http.HandlerFunc(attendanceHandler.GetAttendance)))

	return mux
}
