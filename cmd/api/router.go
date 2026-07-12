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
	"erp-constructora/internal/suppliers"
	"erp-constructora/internal/users"
)

func SetupRoutes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

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
	notificationsService := notifications.NewService(notificationsRepository)
	notificationsHandler := notifications.NewHandler(notificationsService)

	auditLogsRepository := audit.NewRepository(db)
	auditLogsService := audit.NewService(auditLogsRepository)
	auditLogsHandler := audit.NewHandler(auditLogsService)

	subscriptionRepo := subscriptions.NewRepository(db)
	subscriptionService := subscriptions.NewService(subscriptionRepo)
	subscriptionHandler := subscriptions.NewHandler(subscriptionService)

	subMiddleware := middlewares.RequireActiveSubscription(subscriptionService)
	auth := middlewares.AuthMiddleware

	mux.HandleFunc("POST /register", userHandler.RegisterCompanyAndAdmin)
	mux.HandleFunc("POST /login", userHandler.Login)

	// --- Projects ---
	mux.Handle("POST /projects", auth(subMiddleware(http.HandlerFunc(projectHandler.Create))))
	mux.Handle("GET /projects", auth(http.HandlerFunc(projectHandler.GetAll)))
	mux.Handle("PUT /projects/{id}", auth(http.HandlerFunc(projectHandler.Update)))
	mux.Handle("DELETE /projects/{id}", auth(http.HandlerFunc(projectHandler.Delete)))

	// --- Clients ---
	mux.Handle("POST /clients", auth(http.HandlerFunc(clientHandler.Create)))
	mux.Handle("GET /clients", auth(http.HandlerFunc(clientHandler.GetAll)))
	mux.Handle("PUT /clients/{id}", auth(http.HandlerFunc(clientHandler.Update)))
	mux.Handle("DELETE /clients/{id}", auth(http.HandlerFunc(clientHandler.Delete)))

	// --- Budgets ---
	mux.Handle("POST /budgets", auth(http.HandlerFunc(budgetHandler.Create)))
	mux.Handle("GET /budgets/{project_id}", auth(http.HandlerFunc(budgetHandler.GetBudgetsByProjectID)))
	mux.Handle("PUT /budgets/{id}", auth(http.HandlerFunc(budgetHandler.Update)))
	mux.Handle("DELETE /budgets/{id}", auth(http.HandlerFunc(budgetHandler.Delete)))

	// --- Expenses ---
	mux.Handle("POST /expenses", auth(http.HandlerFunc(expenseHandler.Create)))
	mux.Handle("GET /expenses/{project_id}", auth(http.HandlerFunc(expenseHandler.GetByProject)))
	mux.Handle("PUT /expenses/{id}", auth(http.HandlerFunc(expenseHandler.Update)))
	mux.Handle("DELETE /expenses/{id}", auth(http.HandlerFunc(expenseHandler.Delete)))

	// --- Purchase Orders ---
	mux.Handle("POST /purcharse", auth(http.HandlerFunc(purcharseHandler.CreatePurchaseOrder)))
	mux.Handle("GET /purcharse/{project_id}", auth(http.HandlerFunc(purcharseHandler.GetOrdersByProject)))
	mux.Handle("PUT /purcharse/{id}", auth(http.HandlerFunc(purcharseHandler.UpdatePurchaseOrder)))
	mux.Handle("DELETE /purcharse/{id}", auth(http.HandlerFunc(purcharseHandler.DeletePurchaseOrder)))

	// --- Suppliers ---
	mux.Handle("POST /supplier", auth(http.HandlerFunc(suppliersHandler.CreateSupplier)))
	mux.Handle("GET /supplier", auth(http.HandlerFunc(suppliersHandler.GetAllSuppliers)))
	mux.Handle("PUT /supplier/{id}", auth(http.HandlerFunc(suppliersHandler.UpdateSupplier)))
	mux.Handle("DELETE /supplier/{id}", auth(http.HandlerFunc(suppliersHandler.DeleteSupplier)))

	// --- Inventory ---
	mux.Handle("POST /materials", auth(http.HandlerFunc(inventoryHandler.CreateMaterial)))
	mux.Handle("GET /materials", auth(http.HandlerFunc(inventoryHandler.GetAllMaterials)))
	mux.Handle("PUT /materials/{id}", auth(http.HandlerFunc(inventoryHandler.UpdateMaterial)))
	mux.Handle("DELETE /materials/{id}", auth(http.HandlerFunc(inventoryHandler.DeleteMaterial)))
	mux.Handle("POST /warehouses", auth(http.HandlerFunc(inventoryHandler.CreateWarehouse)))
	mux.Handle("PUT /warehouses/{id}", auth(http.HandlerFunc(inventoryHandler.UpdateWarehouse)))
	mux.Handle("DELETE /warehouses/{id}", auth(http.HandlerFunc(inventoryHandler.DeleteWarehouse)))
	mux.Handle("POST /inventory/movements", auth(http.HandlerFunc(inventoryHandler.PostMovement)))
	mux.Handle("GET /inventory/stock/{warehouse_id}", auth(http.HandlerFunc(inventoryHandler.GetStock)))

	// --- Equipment ---
	mux.Handle("POST /equipment", auth(http.HandlerFunc(equipementHandler.CreateEquipment)))
	mux.Handle("GET /equipment", auth(http.HandlerFunc(equipementHandler.GetAll)))
	mux.Handle("PUT /equipment/{id}", auth(http.HandlerFunc(equipementHandler.UpdateEquipment)))
	mux.Handle("DELETE /equipment/{id}", auth(http.HandlerFunc(equipementHandler.DeleteEquipment)))
	mux.Handle("POST /equipment/assignments", auth(http.HandlerFunc(equipementHandler.Assign)))
	mux.Handle("POST /equipment/maintenances", auth(http.HandlerFunc(equipementHandler.Maintenance)))

	// --- Personnel ---
	mux.Handle("POST /positions", auth(http.HandlerFunc(personnelHandler.CreatePosition)))
	mux.Handle("PUT /positions/{id}", auth(http.HandlerFunc(personnelHandler.UpdatePosition)))
	mux.Handle("DELETE /positions/{id}", auth(http.HandlerFunc(personnelHandler.DeletePosition)))
	mux.Handle("POST /employees", auth(http.HandlerFunc(personnelHandler.CreateEmployee)))
	mux.Handle("GET /employees", auth(http.HandlerFunc(personnelHandler.GetEmployees)))
	mux.Handle("PUT /employees/{id}", auth(http.HandlerFunc(personnelHandler.UpdateEmployee)))
	mux.Handle("DELETE /employees/{id}", auth(http.HandlerFunc(personnelHandler.DeleteEmployee)))
	mux.Handle("POST /contracts", auth(http.HandlerFunc(personnelHandler.CreateContract)))
	mux.Handle("PUT /contracts/{id}", auth(http.HandlerFunc(personnelHandler.UpdateContract)))
	mux.Handle("DELETE /contracts/{id}", auth(http.HandlerFunc(personnelHandler.DeleteContract)))

	// --- Attendance ---
	mux.Handle("POST /attendance", auth(http.HandlerFunc(attendanceHandler.SaveAttendance)))
	mux.Handle("GET /attendance/{project_id}", auth(http.HandlerFunc(attendanceHandler.GetAttendance)))
	mux.Handle("PUT /attendance/logs/{id}", auth(http.HandlerFunc(attendanceHandler.UpdateAttendanceLog)))
	mux.Handle("DELETE /attendance/{id}", auth(http.HandlerFunc(attendanceHandler.DeleteAttendance)))

	// --- Contractors ---
	mux.Handle("POST /contractors", auth(http.HandlerFunc(contractorsHandler.CreateContractor)))
	mux.Handle("PUT /contractors/{id}", auth(http.HandlerFunc(contractorsHandler.UpdateContractor)))
	mux.Handle("DELETE /contractors/{id}", auth(http.HandlerFunc(contractorsHandler.DeleteContractor)))
	mux.Handle("POST /contractors/contracts", auth(http.HandlerFunc(contractorsHandler.CreateContract)))
	mux.Handle("PUT /contractors/contracts/{id}", auth(http.HandlerFunc(contractorsHandler.UpdateContractorContract)))
	mux.Handle("DELETE /contractors/contracts/{id}", auth(http.HandlerFunc(contractorsHandler.DeleteContractorContract)))
	mux.Handle("POST /contractors/payments", auth(http.HandlerFunc(contractorsHandler.PostPayment)))
	mux.Handle("GET /contractors/contracts/{project_id}", auth(http.HandlerFunc(contractorsHandler.GetContracts)))

	// --- Schedule ---
	mux.Handle("POST /schedule/tasks", auth(http.HandlerFunc(sheduleHandler.CreateTask)))
	mux.Handle("PUT /schedule/tasks/{id}", auth(http.HandlerFunc(sheduleHandler.UpdateTask)))
	mux.Handle("DELETE /schedule/tasks/{id}", auth(http.HandlerFunc(sheduleHandler.DeleteTask)))
	mux.Handle("POST /schedule/dependencies", auth(http.HandlerFunc(sheduleHandler.AddDependency)))
	mux.Handle("POST /schedule/milestones", auth(http.HandlerFunc(sheduleHandler.CreateMilestone)))
	mux.Handle("PUT /schedule/milestones/{id}", auth(http.HandlerFunc(sheduleHandler.UpdateMilestone)))
	mux.Handle("DELETE /schedule/milestones/{id}", auth(http.HandlerFunc(sheduleHandler.DeleteMilestone)))
	mux.Handle("GET /schedule/{project_id}", auth(http.HandlerFunc(sheduleHandler.GetSchedule)))

	// --- Progress ---
	mux.Handle("POST /progress/daily", auth(http.HandlerFunc(progressHandler.CreateDailyReport)))
	mux.Handle("PUT /progress/daily/{id}", auth(http.HandlerFunc(progressHandler.UpdateDailyReport)))
	mux.Handle("DELETE /progress/daily/{id}", auth(http.HandlerFunc(progressHandler.DeleteDailyReport)))
	mux.Handle("GET /progress/{project_id}", auth(http.HandlerFunc(progressHandler.GetDailyReport)))

	// --- Photos ---
	mux.Handle("POST /photos", auth(http.HandlerFunc(photosHandler.UploadPhotoMetadata)))
	mux.Handle("PUT /photos/{id}", auth(http.HandlerFunc(photosHandler.UpdatePhoto)))
	mux.Handle("DELETE /photos/{id}", auth(http.HandlerFunc(photosHandler.DeletePhoto)))
	mux.Handle("GET /photos/{project_id}", auth(http.HandlerFunc(photosHandler.GetGallery)))

	// --- Invoices / Payments ---
	mux.Handle("POST /invoices", auth(http.HandlerFunc(paymentHandler.CreateInvoice)))
	mux.Handle("PUT /invoices/{id}", auth(http.HandlerFunc(paymentHandler.UpdateInvoice)))
	mux.Handle("DELETE /invoices/{id}", auth(http.HandlerFunc(paymentHandler.DeleteInvoice)))
	mux.Handle("PATCH /invoices/{id}/cancel", auth(http.HandlerFunc(paymentHandler.CancelInvoice)))
	mux.Handle("POST /invoices/payments", auth(http.HandlerFunc(paymentHandler.PostPayment)))

	// --- Dashboard ---
	mux.Handle("GET /dashboard/financial/{project_id}", auth(http.HandlerFunc(dashboardHandler.GetSummary)))

	// --- Documents ---
	mux.Handle("POST /documents/types", auth(http.HandlerFunc(documentsHandler.CreateType)))
	mux.Handle("PUT /documents/types/{id}", auth(http.HandlerFunc(documentsHandler.UpdateDocumentType)))
	mux.Handle("DELETE /documents/types/{id}", auth(http.HandlerFunc(documentsHandler.DeleteDocumentType)))
	mux.Handle("POST /documents", auth(http.HandlerFunc(documentsHandler.CreateDocument)))
	mux.Handle("PUT /documents/{id}", auth(http.HandlerFunc(documentsHandler.UpdateDocument)))
	mux.Handle("DELETE /documents/{id}", auth(http.HandlerFunc(documentsHandler.DeleteDocument)))
	mux.Handle("POST /documents/versions", auth(http.HandlerFunc(documentsHandler.UpdateVersion)))

	// --- Notifications ---
	mux.Handle("POST /notifications", auth(http.HandlerFunc(notificationsHandler.CreateNotifications)))
	mux.Handle("GET /notifications", auth(http.HandlerFunc(notificationsHandler.GetMyNotifications)))
	mux.Handle("PATCH /notifications/{notification_id}/read", auth(http.HandlerFunc(notificationsHandler.MarkRead)))
	mux.Handle("DELETE /notifications/{notification_id}", auth(http.HandlerFunc(notificationsHandler.DeleteNotification)))

	// --- Audit Logs ---
	mux.Handle("POST /audits-logs", auth(http.HandlerFunc(auditLogsHandler.CreateLog)))
	mux.Handle("GET /audits-logs", auth(http.HandlerFunc(auditLogsHandler.GetCompanyLogs)))

	// --- Subscriptions ---
	mux.Handle("GET /subscriptions/me", auth(http.HandlerFunc(subscriptionHandler.GetMySubscription)))
	mux.Handle("POST /subscriptions", auth(http.HandlerFunc(subscriptionHandler.CreateSubscription)))
	mux.Handle("PATCH /subscriptions/{id}", auth(http.HandlerFunc(subscriptionHandler.UpdateSubscription)))

	return mux
}
