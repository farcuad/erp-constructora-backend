package main

import (
	"database/sql"
	"net/http"

	"erp-constructora/internal/attendance"
	audit "erp-constructora/internal/audit_logs"
	"erp-constructora/internal/budgets"
	"erp-constructora/internal/clients"
	"erp-constructora/internal/contractors"
	"erp-constructora/internal/documents"
	"erp-constructora/internal/equipement"
	"erp-constructora/internal/expense"
	financialdashboard "erp-constructora/internal/financial_dashboard"
	"erp-constructora/internal/inventory"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/notifications"
	"erp-constructora/internal/payments"
	"erp-constructora/internal/personnel"
	"erp-constructora/internal/photos"
	"erp-constructora/internal/progress"
	"erp-constructora/internal/project"
	"erp-constructora/internal/purchase"
	schedule "erp-constructora/internal/shedule"
	"erp-constructora/internal/subscriptions"
	"erp-constructora/internal/superadmin"
	"erp-constructora/internal/suppliers"
	"erp-constructora/internal/users"
)

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func SetupRoutes(db *sql.DB) http.Handler {

	mux := http.NewServeMux()

	userRepo := users.NewRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)

	subscriptionRepo := subscriptions.NewRepository(db)
	subscriptionService := subscriptions.NewService(subscriptionRepo)

	projectRepo := project.NewRepository(db)
	projectService := project.NewService(projectRepo, subscriptionService)
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

	contractorsRepository := contractors.NewRepository(db)
	contractorsService := contractors.NewService(contractorsRepository)
	contractorsHandler := contractors.NewHandler(contractorsService)

	sheduleRepository := schedule.NewRepository(db)
	sheduleService := schedule.NewService(sheduleRepository)
	sheduleHandler := schedule.NewHandler(sheduleService)

	progressRepository := progress.NewRepository(db)
	progressService := progress.NewService(progressRepository)
	progressHandler := progress.NewHandler(progressService)

	photosRepository := photos.NewRepository(db)
	photosService := photos.NewService(photosRepository)
	photosHandler := photos.NewHandler(photosService)

	paymentRepository := payments.NewRepository(db)
	paymentService := payments.NewService(paymentRepository)
	paymentHandler := payments.NewHandler(paymentService)

	dashboardRepository := financialdashboard.NewRepository(db)
	dashboardService := financialdashboard.NewService(dashboardRepository)
	dashboardHandler := financialdashboard.NewHandler(dashboardService)

	documentsRepository := documents.NewRepository(db)
	documentsService := documents.NewService(documentsRepository)
	documentsHandler := documents.NewHandler(documentsService)

	notificationsRepository := notifications.NewRepository(db)
	notificationsHub := notifications.NewWSHub()

	// 2. Inyectamos el Hub tanto al Service (para enviar eventos) como al Handler (para la conexión WS)
	notificationsService := notifications.NewService(notificationsRepository, notificationsHub)
	notificationsHandler := notifications.NewHandler(notificationsService, notificationsHub)

	auditLogsRepository := audit.NewRepository(db)
	auditLogsService := audit.NewService(auditLogsRepository)
	auditLogsHandler := audit.NewHandler(auditLogsService)

	subscriptionHandler := subscriptions.NewHandler(subscriptionService)

	superAdminRepo := superadmin.NewRepository(db)
	superAdminService := superadmin.NewService(superAdminRepo)
	superAdminHandler := superadmin.NewHandler(superAdminService)

	subMiddleware := middlewares.RequireActiveSubscription(subscriptionService)
	auth := middlewares.AuthMiddleware
	adminOnly := middlewares.RequireSuperAdmin

	// protected: valida token + suscripción activa + permiso del JWT
	protected := func(perm string, h http.HandlerFunc) http.Handler {
		return chain(http.HandlerFunc(h), auth, subMiddleware, middlewares.RequirePermission(perm))
	}

	// protectedBasic: valida solo token + permiso (sin suscripción)
	protectedBasic := func(perm string, h http.HandlerFunc) http.Handler {
		return chain(http.HandlerFunc(h), auth, middlewares.RequirePermission(perm))
	}

	// --- Public Routes ---

	mux.HandleFunc("POST /register", userHandler.RegisterCompanyAndAdmin)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /admin/login", superAdminHandler.Login)

	// --- Users & Roles ---
	mux.Handle("GET /roles", protected("users:read", userHandler.GetRoles))
	mux.Handle("GET /users", protected("users:read", userHandler.GetUsers))
	mux.Handle("POST /users", protected("users:create", userHandler.CreateUser))
	mux.Handle("PUT /users/{id}", protected("users:update", userHandler.UpdateUser))
	mux.Handle("DELETE /users/{id}", protected("users:delete", userHandler.DeleteUser))

	// --- Projects ---
	mux.Handle("POST /projects", protected("projects:create", projectHandler.Create))
	mux.Handle("GET /projects", protected("projects:read", projectHandler.GetAll))
	mux.Handle("PUT /projects/{id}", protected("projects:update", projectHandler.Update))
	mux.Handle("DELETE /projects/{id}", protected("projects:delete", projectHandler.Delete))

	// --- Clients ---
	mux.Handle("POST /clients", protected("clients:create", clientHandler.Create))
	mux.Handle("GET /clients", protected("clients:read", clientHandler.GetAll))
	mux.Handle("PUT /clients/{id}", protected("clients:update", clientHandler.Update))
	mux.Handle("DELETE /clients/{id}", protected("clients:delete", clientHandler.Delete))

	// --- Budgets ---
	mux.Handle("POST /budgets", protected("budgets:create", budgetHandler.Create))
	mux.Handle("GET /budgets/{project_id}", protected("budgets:read", budgetHandler.GetBudgetsByProjectID))
	mux.Handle("PUT /budgets/{id}", protected("budgets:update", budgetHandler.Update))
	mux.Handle("DELETE /budgets/{id}", protected("budgets:delete", budgetHandler.Delete))

	// --- Expenses ---
	mux.Handle("POST /expenses", protected("expenses:create", expenseHandler.Create))
	mux.Handle("GET /expenses/{project_id}", protected("expenses:read", expenseHandler.GetByProject))
	mux.Handle("PUT /expenses/{id}", protected("expenses:update", expenseHandler.Update))
	mux.Handle("DELETE /expenses/{id}", protected("expenses:delete", expenseHandler.Delete))

	// --- Purchase Orders ---
	mux.Handle("POST /purcharse", protected("purchases:create", purcharseHandler.CreatePurchaseOrder))
	mux.Handle("GET /purcharse/{project_id}", protected("purchases:read", purcharseHandler.GetOrdersByProject))
	mux.Handle("PUT /purcharse/{id}", protected("purchases:update", purcharseHandler.UpdatePurchaseOrder))
	mux.Handle("DELETE /purcharse/{id}", protected("purchases:delete", purcharseHandler.DeletePurchaseOrder))

	// --- Suppliers ---
	mux.Handle("POST /supplier", protected("suppliers:create", suppliersHandler.CreateSupplier))
	mux.Handle("GET /supplier", protected("suppliers:read", suppliersHandler.GetAllSuppliers))
	mux.Handle("PUT /supplier/{id}", protected("suppliers:update", suppliersHandler.UpdateSupplier))
	mux.Handle("DELETE /supplier/{id}", protected("suppliers:delete", suppliersHandler.DeleteSupplier))

	// --- Inventory ---
	mux.Handle("POST /materials", protected("inventory:manage", inventoryHandler.CreateMaterial))
	mux.Handle("GET /materials", protected("inventory:read", inventoryHandler.GetAllMaterials))
	mux.Handle("PUT /materials/{id}", protected("inventory:manage", inventoryHandler.UpdateMaterial))
	mux.Handle("DELETE /materials/{id}", protected("inventory:manage", inventoryHandler.DeleteMaterial))
	mux.Handle("POST /warehouses", protected("inventory:manage", inventoryHandler.CreateWarehouse))
	mux.Handle("GET /warehouses", protected("inventory:read", inventoryHandler.GetAllWarehouses))
	mux.Handle("PUT /warehouses/{id}", protected("inventory:manage", inventoryHandler.UpdateWarehouse))
	mux.Handle("DELETE /warehouses/{id}", protected("inventory:manage", inventoryHandler.DeleteWarehouse))
	mux.Handle("POST /inventory/movements", protected("inventory:manage", inventoryHandler.PostMovement))
	mux.Handle("GET /inventory/stock/{warehouse_id}", protected("inventory:read", inventoryHandler.GetStock))

	// --- Equipment ---
	mux.Handle("POST /equipment/types", protected("equipment:manage", equipementHandler.CreateEquipmentType))
	mux.Handle("GET /equipment/types", protected("equipment:read", equipementHandler.GetAllEquipmentTypes))
	mux.Handle("POST /equipment", protected("equipment:manage", equipementHandler.CreateEquipment))
	mux.Handle("GET /equipment", protected("equipment:read", equipementHandler.GetAll))
	mux.Handle("PUT /equipment/{id}", protected("equipment:manage", equipementHandler.UpdateEquipment))
	mux.Handle("DELETE /equipment/{id}", protected("equipment:manage", equipementHandler.DeleteEquipment))
	mux.Handle("POST /equipment/assignments", protected("equipment:assign", equipementHandler.Assign))
	mux.Handle("GET /equipment/assignments/{equipment_id}", protected("equipment:read", equipementHandler.GetAssignment))
	mux.Handle("POST /equipment/maintenances", protected("equipment:manage", equipementHandler.Maintenance))
	mux.Handle("GET /equipment/maintenances/{equipment_id}", protected("equipment:read", equipementHandler.GetMaintenanceById))

	// --- Personnel ---
	mux.Handle("POST /positions", protected("personnel:manage", personnelHandler.CreatePosition))
	mux.Handle("GET /positions", protected("personnel:read", personnelHandler.GetAllPositions))
	mux.Handle("PUT /positions/{id}", protected("personnel:manage", personnelHandler.UpdatePosition))
	mux.Handle("DELETE /positions/{id}", protected("personnel:manage", personnelHandler.DeletePosition))
	mux.Handle("POST /employees", protected("personnel:manage", personnelHandler.CreateEmployee))
	mux.Handle("GET /employees", protected("personnel:read", personnelHandler.GetEmployees))
	mux.Handle("PUT /employees/{id}", protected("personnel:manage", personnelHandler.UpdateEmployee))
	mux.Handle("DELETE /employees/{id}", protected("personnel:manage", personnelHandler.DeleteEmployee))
	mux.Handle("POST /contracts", protected("personnel:manage", personnelHandler.CreateContract))
	mux.Handle("GET /contracts/{project_id}", protected("personnel:read", personnelHandler.GetALlContracts))
	mux.Handle("PUT /contracts/{id}", protected("personnel:manage", personnelHandler.UpdateContract))
	mux.Handle("DELETE /contracts/{id}", protected("personnel:manage", personnelHandler.DeleteContract))

	// --- Attendance ---
	mux.Handle("POST /attendance", protected("attendance:mark", attendanceHandler.SaveAttendance))
	mux.Handle("GET /attendance/{project_id}", protected("attendance:read", attendanceHandler.GetAttendance))
	mux.Handle("PUT /attendance/logs/{id}", protected("attendance:mark", attendanceHandler.UpdateAttendanceLog))
	mux.Handle("DELETE /attendance/{id}", protected("attendance:mark", attendanceHandler.DeleteAttendance))

	// --- Contractors ---
	mux.Handle("POST /contractors", protected("contractors:manage", contractorsHandler.CreateContractor))
	mux.Handle("GET /contractors", protected("contractors:read", contractorsHandler.GetALlContracts))
	mux.Handle("PUT /contractors/{id}", protected("contractors:manage", contractorsHandler.UpdateContractor))
	mux.Handle("DELETE /contractors/{id}", protected("contractors:manage", contractorsHandler.DeleteContractor))
	mux.Handle("POST /contractors/contracts", protected("contractors:manage", contractorsHandler.CreateContract))
	mux.Handle("GET /contractors/contracts/{project_id}", protected("contractors:read", contractorsHandler.GetContracts))
	mux.Handle("PUT /contractors/contracts/{id}", protected("contractors:manage", contractorsHandler.UpdateContractorContract))
	mux.Handle("DELETE /contractors/contracts/{id}", protected("contractors:manage", contractorsHandler.DeleteContractorContract))
	mux.Handle("POST /contractors/payments", protected("contractors:pay", contractorsHandler.PostPayment))
	mux.Handle("GET /contractors/payments", protected("contractors:read", contractorsHandler.GetAllContractPayments))

	// --- Schedule ---
	mux.Handle("POST /schedule/tasks", protected("schedule:update", sheduleHandler.CreateTask))
	mux.Handle("PUT /schedule/tasks/{id}", protected("schedule:update", sheduleHandler.UpdateTask))
	mux.Handle("DELETE /schedule/tasks/{id}", protected("schedule:update", sheduleHandler.DeleteTask))
	mux.Handle("GET /schedule/{project_id}", protected("schedule:read", sheduleHandler.GetSchedule))

	// --- Progress ---
	mux.Handle("POST /progress/daily", protected("progress:create", progressHandler.CreateDailyReport))
	mux.Handle("PUT /progress/daily/{id}", protected("progress:update", progressHandler.UpdateDailyReport))
	mux.Handle("DELETE /progress/daily/{id}", protected("progress:delete", progressHandler.DeleteDailyReport))
	mux.Handle("GET /progress/{project_id}", protected("progress:read", progressHandler.GetDailyReport))

	// --- Photos ---
	mux.Handle("POST /photos", protected("photos:upload", photosHandler.UploadPhotoMetadata))
	mux.Handle("PUT /photos/{id}", protected("photos:upload", photosHandler.UpdatePhoto))
	mux.Handle("DELETE /photos/{id}", protected("photos:delete", photosHandler.DeletePhoto))
	mux.Handle("GET /photos/{project_id}", protected("photos:read", photosHandler.GetGallery))

	// --- Invoices / Payments ---
	mux.Handle("GET /invoices/{id}", protected("invoices:read", paymentHandler.GetInvoiceByID))
	mux.Handle("GET /invoices/project/{project_id}", protected("invoices:read", paymentHandler.GetInvoices))
	mux.Handle("POST /invoices", protected("invoices:create", paymentHandler.CreateInvoice))
	mux.Handle("PUT /invoices/{id}", protected("invoices:update", paymentHandler.UpdateInvoice))
	mux.Handle("DELETE /invoices/{id}", protected("invoices:delete", paymentHandler.DeleteInvoice))
	mux.Handle("PATCH /invoices/{id}/cancel", protected("invoices:cancel", paymentHandler.CancelInvoice))
	mux.Handle("GET /invoices/payments/{invoice_id}", protected("invoices:read", paymentHandler.GetPayments))
	mux.Handle("POST /invoices/payments", protected("invoices:pay", paymentHandler.PostPayment))

	// --- Dashboard ---
	mux.Handle("GET /dashboard/financial/{project_id}", protected("dashboard:read", dashboardHandler.GetSummary))

	// --- Documents ---
	mux.Handle("GET /documents/types", protected("documents:read", documentsHandler.GetTypes))
	mux.Handle("POST /documents/types", protected("documents:create", documentsHandler.CreateType))
	mux.Handle("PUT /documents/types/{id}", protected("documents:update", documentsHandler.UpdateDocumentType))
	mux.Handle("DELETE /documents/types/{id}", protected("documents:delete", documentsHandler.DeleteDocumentType))
	mux.Handle("GET /documents/{id}", protected("documents:read", documentsHandler.GetDocumentByID))
	mux.Handle("GET /documents/project/{project_id}", protected("documents:read", documentsHandler.GetDocuments))
	mux.Handle("POST /documents", protected("documents:create", documentsHandler.CreateDocument))
	mux.Handle("PUT /documents/{id}", protected("documents:update", documentsHandler.UpdateDocument))
	mux.Handle("DELETE /documents/{id}", protected("documents:delete", documentsHandler.DeleteDocument))
	mux.Handle("GET /documents/versions/{document_id}", protected("documents:read", documentsHandler.GetVersions))
	mux.Handle("POST /documents/versions", protected("documents:update", documentsHandler.UpdateVersion))

	// --- Notifications ---
	mux.Handle("GET /notifications/ws", protectedBasic("notifications:read", notificationsHandler.HandleWS))
	mux.Handle("POST /notifications", protectedBasic("notifications:manage", notificationsHandler.CreateNotifications))
	mux.Handle("GET /notifications", protectedBasic("notifications:read", notificationsHandler.GetMyNotifications))
	mux.Handle("PATCH /notifications/{notification_id}/read", protectedBasic("notifications:read", notificationsHandler.MarkRead))
	mux.Handle("DELETE /notifications/{notification_id}", protectedBasic("notifications:manage", notificationsHandler.DeleteNotification))

	// --- Audit Logs ---
	mux.Handle("POST /audits-logs", protectedBasic("audits:read", auditLogsHandler.CreateLog))
	mux.Handle("GET /audits-logs", protectedBasic("audits:read", auditLogsHandler.GetCompanyLogs))

	// --- Subscriptions ---
	mux.Handle("GET /subscriptions/me", chain(http.HandlerFunc(subscriptionHandler.GetMySubscription), auth))
	mux.Handle("GET /subscriptions", chain(http.HandlerFunc(subscriptionHandler.GetAllSubscriptions), adminOnly, auth))
	mux.Handle("GET /subscriptions/{id}", chain(http.HandlerFunc(subscriptionHandler.GetSubscriptionByID), adminOnly, auth))
	mux.Handle("POST /subscriptions", chain(http.HandlerFunc(subscriptionHandler.CreateSubscription), adminOnly, auth))
	mux.Handle("PATCH /subscriptions/{id}", chain(http.HandlerFunc(subscriptionHandler.UpdateSubscription), adminOnly, auth))

	return mux
}
