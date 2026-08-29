package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"erp-constructora/internal/app"
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
	"erp-constructora/internal/rag"
	"erp-constructora/internal/rag/worker"
	schedule "erp-constructora/internal/shedule"
	"erp-constructora/internal/subscriptions"
	"erp-constructora/internal/superadmin"
	"erp-constructora/internal/suppliers"
	"erp-constructora/internal/users"
	"erp-constructora/pkg/fcm"
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

	notificationsRepository := notifications.NewRepository(db)
	notificationsHub := notifications.NewWSHub()

	// Cliente FCM para push notifications (opcional: si no hay credenciales, solo se usan WebSockets).
	// Prioridad: 1) FIREBASE_CREDENTIALS_JSON (JSON crudo o Base64, para VPS/contenedores)
	//            2) FIREBASE_CREDENTIALS_PATH o config/config-service-account.json (desarrollo local)
	var pushSender notifications.PushSender
	var fcmClient *fcm.FCMClient
	var fcmErr error
	if credsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON"); credsJSON != "" {
		trimmed := strings.TrimSpace(credsJSON)

		var data []byte
		if strings.HasPrefix(trimmed, "{") {
			data = []byte(trimmed)
			log.Printf("Diagnóstico FCM: JSON recibido directo (%d bytes)", len(trimmed))
		} else {
			decoded, err := base64.StdEncoding.DecodeString(trimmed)
			if err != nil {
				log.Printf("Aviso: FIREBASE_CREDENTIALS_JSON no empieza con '{' y tampoco es Base64 válido: %v", err)
			}
			data = decoded
			log.Printf("Diagnóstico FCM: Base64 recibido (%d chars)", len(trimmed))
		}

		fcmClient, fcmErr = fcm.NewFCMClientFromJSON(data, os.Getenv("FIREBASE_PROJECT_ID"))
	} else {
		fcmPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		if fcmPath == "" {
			fcmPath = "config/config-service-account.json"
		}
		fcmClient, fcmErr = fcm.NewFCMClient(fcmPath)
	}
	if fcmErr != nil {
		log.Printf("Aviso: FCM deshabilitado (%v). Solo notificaciones por WebSocket.", fcmErr)
	} else {
		pushSender = fcmClient
		log.Println("Cliente FCM inicializado correctamente")
	}

	// Inyectamos el Hub tanto al Service (para enviar eventos) como al Handler (para la conexión WS)
	notificationsService := notifications.NewService(notificationsRepository, notificationsHub, pushSender)
	notificationsHandler := notifications.NewHandler(notificationsService, notificationsHub)

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
	budgetHandler := budgets.NewHandler(budgetService, notificationsService)

	expenseRepo := expense.NewRepository(db)
	expenseService := expense.NewService(expenseRepo)
	expenseHandler := expense.NewHandler(expenseService, notificationsService)

	purchaseRepo := purchase.NewRepository(db)
	purchaseService := purchase.NewService(purchaseRepo)
	purcharseHandler := purchase.NewHandler(purchaseService, notificationsService)

	supplierRepo := suppliers.NewRepository(db)
	suppilerService := suppliers.NewService(supplierRepo)
	suppliersHandler := suppliers.NewHandler(suppilerService)

	inventoryRepository := inventory.NewRepository(db)
	inventoryService := inventory.NewService(inventoryRepository)
	inventoryHandler := inventory.NewHandler(inventoryService)

	equipementRepository := equipement.NewRepository(db)
	equipementService := equipement.NewService(equipementRepository)
	equipementHandler := equipement.NewHandler(equipementService, notificationsService)

	personnelRespository := personnel.NewRepository(db)
	personnelService := personnel.NewService(personnelRespository)
	personnelHandler := personnel.NewHandler(personnelService)

	attendanceRepositoy := attendance.NewRepository(db)
	attendanceService := attendance.NewService(attendanceRepositoy)
	attendanceHandler := attendance.NewHandler(attendanceService, notificationsService)

	contractorsRepository := contractors.NewRepository(db)
	contractorsService := contractors.NewService(contractorsRepository)
	contractorsHandler := contractors.NewHandler(contractorsService, notificationsService)

	sheduleRepository := schedule.NewRepository(db)
	sheduleService := schedule.NewService(sheduleRepository)
	sheduleHandler := schedule.NewHandler(sheduleService, notificationsService)

	progressRepository := progress.NewRepository(db)
	progressService := progress.NewService(progressRepository)
	progressHandler := progress.NewHandler(progressService, notificationsService)

	photosRepository := photos.NewRepository(db)
	photosService := photos.NewService(photosRepository)
	photosHandler := photos.NewHandler(photosService)

	paymentRepository := payments.NewRepository(db)
	paymentService := payments.NewService(paymentRepository)
	paymentHandler := payments.NewHandler(paymentService, notificationsService)

	dashboardRepository := financialdashboard.NewRepository(db)
	dashboardService := financialdashboard.NewService(dashboardRepository)
	dashboardHandler := financialdashboard.NewHandler(dashboardService)

	appRepo := app.NewRepository(db)
	appService := app.NewService(appRepo)
	appHandler := app.NewHandler(appService)

	openAIKey := os.Getenv("OPENAI_API_KEY")
	chatModel := os.Getenv("OPENAI_CHAT_MODEL")

	ragEmbedder := worker.NewOpenAIEmbedder(openAIKey)
	ragRepository := rag.NewRepository(db)
	ragService := rag.NewService(ragRepository, ragEmbedder, openAIKey, chatModel)
	ragHandler := rag.NewHandler(ragService)

	documentsRepository := documents.NewRepository(db)
	documentsService := documents.NewService(documentsRepository, ragService)
	documentsHandler := documents.NewHandler(documentsService, notificationsService)

	auditLogsRepository := audit.NewRepository(db)
	auditLogsService := audit.NewService(auditLogsRepository)
	auditLogsHandler := audit.NewHandler(auditLogsService, notificationsService)

	subscriptionHandler := subscriptions.NewHandler(subscriptionService)

	superAdminRepo := superadmin.NewRepository(db)
	superAdminService := superadmin.NewService(superAdminRepo)
	superAdminHandler := superadmin.NewHandler(superAdminService)

	subMiddleware := middlewares.RequireActiveSubscription(subscriptionService)
	auth := middlewares.AuthMiddleware
	adminOnly := middlewares.RequireSuperAdmin

	// Grupos de roles para proteger las rutas. El rol "Administrador" siempre tiene acceso total.
	managerRoles := []string{"Gerente"} // gestión general: usuarios, clientes, altas/bajas
	siteRoles := []string{"Gerente", "Ingeniero", "Supervisor"}
	sitePlanning := []string{"Gerente", "Ingeniero"}
	purchaseRoles := []string{"Gerente", "Compras"}
	warehouseRoles := []string{"Gerente", "Almacén"}
	warehouseSite := []string{"Gerente", "Ingeniero", "Almacén"}
	financeRoles := []string{"Gerente", "Contabilidad"}
	siteAndFinance := []string{"Gerente", "Ingeniero", "Supervisor", "Contabilidad"}
	allRoles := []string{"Gerente", "Ingeniero", "Supervisor", "Compras", "Contabilidad", "Almacén"}

	// protected: valida token + suscripción activa + rol del JWT
	protected := func(roles []string, h http.HandlerFunc) http.Handler {
		return chain(http.HandlerFunc(h), auth, subMiddleware, middlewares.RequireRole(roles...))
	}

	// protectedBasic: valida solo token + rol (sin suscripción)
	protectedBasic := func(roles []string, h http.HandlerFunc) http.Handler {
		return chain(http.HandlerFunc(h), auth, middlewares.RequireRole(roles...))
	}

	// --- Public Routes ---

	mux.HandleFunc("POST /register", userHandler.RegisterCompanyAndAdmin)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /admin/login", superAdminHandler.Login)

	// --- Users & Roles ---
	mux.Handle("GET /roles", protected(managerRoles, userHandler.GetRoles))
	mux.Handle("GET /users", protected(managerRoles, userHandler.GetUsers))
	mux.Handle("POST /users", protected(managerRoles, userHandler.CreateUser))
	mux.Handle("PUT /users/{id}", protected(managerRoles, userHandler.UpdateUser))
	mux.Handle("DELETE /users/{id}", protected(managerRoles, userHandler.DeleteUser))

	// --- Projects ---
	mux.Handle("POST /projects", protected(managerRoles, projectHandler.Create))
	mux.Handle("GET /projects", protected(allRoles, projectHandler.GetAll))
	mux.Handle("PUT /projects/{id}", protected(managerRoles, projectHandler.Update))
	mux.Handle("DELETE /projects/{id}", protected(managerRoles, projectHandler.Delete))

	// --- Clients ---
	mux.Handle("POST /clients", protected(managerRoles, clientHandler.Create))
	mux.Handle("GET /clients", protected(allRoles, clientHandler.GetAll))
	mux.Handle("PUT /clients/{id}", protected(managerRoles, clientHandler.Update))
	mux.Handle("DELETE /clients/{id}", protected(managerRoles, clientHandler.Delete))

	// --- Budgets ---
	mux.Handle("POST /budgets", protected(managerRoles, budgetHandler.Create))
	mux.Handle("GET /budgets/{project_id}", protected(siteAndFinance, budgetHandler.GetBudgetsByProjectID))
	mux.Handle("PUT /budgets/{id}", protected(managerRoles, budgetHandler.Update))
	mux.Handle("DELETE /budgets/{id}", protected(managerRoles, budgetHandler.Delete))

	// --- Expenses ---
	mux.Handle("POST /expenses", protected(siteAndFinance, expenseHandler.Create))
	mux.Handle("GET /expenses/{project_id}", protected(siteAndFinance, expenseHandler.GetByProject))
	mux.Handle("PUT /expenses/{id}", protected(financeRoles, expenseHandler.Update))
	mux.Handle("DELETE /expenses/{id}", protected(managerRoles, expenseHandler.Delete))

	// --- Purchase Orders ---
	mux.Handle("POST /purcharse", protected(purchaseRoles, purcharseHandler.CreatePurchaseOrder))
	mux.Handle("GET /purcharse/{project_id}", protected(allRoles, purcharseHandler.GetOrdersByProject))
	mux.Handle("PUT /purcharse/{id}", protected(purchaseRoles, purcharseHandler.UpdatePurchaseOrder))
	mux.Handle("DELETE /purcharse/{id}", protected(managerRoles, purcharseHandler.DeletePurchaseOrder))

	// --- Suppliers ---
	mux.Handle("POST /supplier", protected(purchaseRoles, suppliersHandler.CreateSupplier))
	mux.Handle("GET /supplier", protected(allRoles, suppliersHandler.GetAllSuppliers))
	mux.Handle("PUT /supplier/{id}", protected(purchaseRoles, suppliersHandler.UpdateSupplier))
	mux.Handle("DELETE /supplier/{id}", protected(managerRoles, suppliersHandler.DeleteSupplier))

	// --- Inventory ---
	mux.Handle("POST /materials", protected(warehouseRoles, inventoryHandler.CreateMaterial))
	mux.Handle("GET /materials", protected(allRoles, inventoryHandler.GetAllMaterials))
	mux.Handle("PUT /materials/{id}", protected(warehouseRoles, inventoryHandler.UpdateMaterial))
	mux.Handle("DELETE /materials/{id}", protected(warehouseRoles, inventoryHandler.DeleteMaterial))
	mux.Handle("POST /warehouses", protected(warehouseRoles, inventoryHandler.CreateWarehouse))
	mux.Handle("GET /warehouses", protected(allRoles, inventoryHandler.GetAllWarehouses))
	mux.Handle("PUT /warehouses/{id}", protected(warehouseRoles, inventoryHandler.UpdateWarehouse))
	mux.Handle("DELETE /warehouses/{id}", protected(warehouseRoles, inventoryHandler.DeleteWarehouse))
	mux.Handle("POST /inventory/movements", protected(warehouseRoles, inventoryHandler.PostMovement))
	mux.Handle("GET /inventory/stock/{warehouse_id}", protected(allRoles, inventoryHandler.GetStock))

	// --- Equipment ---
	mux.Handle("POST /equipment/types", protected(warehouseRoles, equipementHandler.CreateEquipmentType))
	mux.Handle("GET /equipment/types", protected(allRoles, equipementHandler.GetAllEquipmentTypes))
	mux.Handle("POST /equipment", protected(warehouseRoles, equipementHandler.CreateEquipment))
	mux.Handle("GET /equipment", protected(allRoles, equipementHandler.GetAll))
	mux.Handle("PUT /equipment/{id}", protected(warehouseRoles, equipementHandler.UpdateEquipment))
	mux.Handle("DELETE /equipment/{id}", protected(managerRoles, equipementHandler.DeleteEquipment))
	mux.Handle("POST /equipment/assignments", protected(siteRoles, equipementHandler.Assign))
	mux.Handle("GET /equipment/assignments/{equipment_id}", protected(allRoles, equipementHandler.GetAssignment))
	mux.Handle("POST /equipment/maintenances", protected(warehouseSite, equipementHandler.Maintenance))
	mux.Handle("GET /equipment/maintenances/{equipment_id}", protected(allRoles, equipementHandler.GetMaintenanceById))

	// --- Personnel ---
	mux.Handle("POST /positions", protected(managerRoles, personnelHandler.CreatePosition))
	mux.Handle("GET /positions", protected(allRoles, personnelHandler.GetAllPositions))
	mux.Handle("PUT /positions/{id}", protected(managerRoles, personnelHandler.UpdatePosition))
	mux.Handle("DELETE /positions/{id}", protected(managerRoles, personnelHandler.DeletePosition))
	mux.Handle("POST /employees", protected(managerRoles, personnelHandler.CreateEmployee))
	mux.Handle("GET /employees", protected(siteAndFinance, personnelHandler.GetEmployees))
	mux.Handle("PUT /employees/{id}", protected(managerRoles, personnelHandler.UpdateEmployee))
	mux.Handle("DELETE /employees/{id}", protected(managerRoles, personnelHandler.DeleteEmployee))
	mux.Handle("POST /contracts", protected(managerRoles, personnelHandler.CreateContract))
	mux.Handle("GET /contracts/{project_id}", protected(siteAndFinance, personnelHandler.GetALlContracts))
	mux.Handle("PUT /contracts/{id}", protected(managerRoles, personnelHandler.UpdateContract))
	mux.Handle("DELETE /contracts/{id}", protected(managerRoles, personnelHandler.DeleteContract))

	// --- Attendance ---
	mux.Handle("POST /attendance", protected(siteRoles, attendanceHandler.SaveAttendance))
	mux.Handle("GET /attendance/{project_id}", protected(siteAndFinance, attendanceHandler.GetAttendance))
	mux.Handle("PUT /attendance/logs/{id}", protected(siteRoles, attendanceHandler.UpdateAttendanceLog))
	mux.Handle("DELETE /attendance/{id}", protected(managerRoles, attendanceHandler.DeleteAttendance))

	// --- Contractors ---
	mux.Handle("POST /contractors", protected(managerRoles, contractorsHandler.CreateContractor))
	mux.Handle("GET /contractors", protected(siteAndFinance, contractorsHandler.GetALlContracts))
	mux.Handle("PUT /contractors/{id}", protected(managerRoles, contractorsHandler.UpdateContractor))
	mux.Handle("DELETE /contractors/{id}", protected(managerRoles, contractorsHandler.DeleteContractor))
	mux.Handle("POST /contractors/contracts", protected(managerRoles, contractorsHandler.CreateContract))
	mux.Handle("GET /contractors/contracts/{project_id}", protected(siteAndFinance, contractorsHandler.GetContracts))
	mux.Handle("PUT /contractors/contracts/{id}", protected(managerRoles, contractorsHandler.UpdateContractorContract))
	mux.Handle("DELETE /contractors/contracts/{id}", protected(managerRoles, contractorsHandler.DeleteContractorContract))
	mux.Handle("POST /contractors/payments", protected(financeRoles, contractorsHandler.PostPayment))
	mux.Handle("GET /contractors/payments", protected(financeRoles, contractorsHandler.GetAllContractPayments))

	// --- Schedule ---
	mux.Handle("POST /schedule/tasks", protected(sitePlanning, sheduleHandler.CreateTask))
	mux.Handle("PUT /schedule/tasks/{id}", protected(sitePlanning, sheduleHandler.UpdateTask))
	mux.Handle("DELETE /schedule/tasks/{id}", protected(managerRoles, sheduleHandler.DeleteTask))
	mux.Handle("GET /schedule/{project_id}", protected(allRoles, sheduleHandler.GetSchedule))

	// --- Progress ---
	mux.Handle("POST /progress/daily", protected(siteRoles, progressHandler.CreateDailyReport))
	mux.Handle("PUT /progress/daily/{id}", protected(siteRoles, progressHandler.UpdateDailyReport))
	mux.Handle("DELETE /progress/daily/{id}", protected(managerRoles, progressHandler.DeleteDailyReport))
	mux.Handle("GET /progress/{project_id}", protected(allRoles, progressHandler.GetDailyReport))

	// --- Photos ---
	mux.Handle("POST /photos", protected(siteRoles, photosHandler.UploadPhotoMetadata))
	mux.Handle("PUT /photos/{id}", protected(siteRoles, photosHandler.UpdatePhoto))
	mux.Handle("DELETE /photos/{id}", protected(managerRoles, photosHandler.DeletePhoto))
	mux.Handle("GET /photos/{project_id}", protected(allRoles, photosHandler.GetGallery))

	// --- Invoices / Payments ---
	mux.Handle("GET /invoices/{id}", protected(financeRoles, paymentHandler.GetInvoiceByID))
	mux.Handle("GET /invoices/project/{project_id}", protected(financeRoles, paymentHandler.GetInvoices))
	mux.Handle("POST /invoices", protected(financeRoles, paymentHandler.CreateInvoice))
	mux.Handle("PUT /invoices/{id}", protected(financeRoles, paymentHandler.UpdateInvoice))
	mux.Handle("DELETE /invoices/{id}", protected(managerRoles, paymentHandler.DeleteInvoice))
	mux.Handle("PATCH /invoices/{id}/cancel", protected(financeRoles, paymentHandler.CancelInvoice))
	mux.Handle("GET /invoices/payments/{invoice_id}", protected(financeRoles, paymentHandler.GetPayments))
	mux.Handle("POST /invoices/payments", protected(financeRoles, paymentHandler.PostPayment))

	// --- Dashboard ---
	mux.Handle("GET /dashboard/financial/{project_id}", protected(financeRoles, dashboardHandler.GetSummary))

	// --- Documents ---
	mux.Handle("GET /documents/types", protected(allRoles, documentsHandler.GetTypes))
	mux.Handle("POST /documents/types", protected(managerRoles, documentsHandler.CreateType))
	mux.Handle("PUT /documents/types/{id}", protected(managerRoles, documentsHandler.UpdateDocumentType))
	mux.Handle("DELETE /documents/types/{id}", protected(managerRoles, documentsHandler.DeleteDocumentType))
	mux.Handle("GET /documents/{id}", protected(allRoles, documentsHandler.GetDocumentByID))
	mux.Handle("GET /documents/project/{project_id}", protected(allRoles, documentsHandler.GetDocuments))
	mux.Handle("POST /documents", protected(allRoles, documentsHandler.CreateDocument))
	mux.Handle("PUT /documents/{id}", protected(allRoles, documentsHandler.UpdateDocument))
	mux.Handle("DELETE /documents/{id}", protected(managerRoles, documentsHandler.DeleteDocument))
	mux.Handle("GET /documents/versions/{document_id}", protected(allRoles, documentsHandler.GetVersions))
	mux.Handle("POST /documents/versions", protected(allRoles, documentsHandler.UpdateVersion))

	// --- Notifications ---
	mux.Handle("GET /notifications/ws", protectedBasic(allRoles, notificationsHandler.HandleWS))
	mux.Handle("GET /notifications", protectedBasic(allRoles, notificationsHandler.GetMyNotifications))
	mux.Handle("PATCH /notifications/{notification_id}/read", protectedBasic(allRoles, notificationsHandler.MarkRead))
	mux.Handle("DELETE /notifications/{notification_id}", protectedBasic(managerRoles, notificationsHandler.DeleteNotification))
	mux.Handle("POST /notifications/push-tokens", protectedBasic(allRoles, notificationsHandler.RegisterPushToken))
	mux.Handle("DELETE /notifications/push-tokens", protectedBasic(allRoles, notificationsHandler.UnregisterPushToken))

	// --- Audit Logs ---
	mux.Handle("POST /audits-logs", protectedBasic(allRoles, auditLogsHandler.CreateLog))
	mux.Handle("GET /audits-logs", protectedBasic(allRoles, auditLogsHandler.GetCompanyLogs))
	mux.Handle("POST /rag/chat", protected(siteRoles, ragHandler.HandleChat))

	// --- App Releases (protegidas por API key, no por JWT) ---
	mux.Handle("POST /app/releases", chain(http.HandlerFunc(appHandler.PublishRelease), middlewares.RequireAPIKey))
	mux.Handle("GET /app/latest", chain(http.HandlerFunc(appHandler.GetLatestRelease), middlewares.RequireAPIKey))

	// --- Subscriptions ---
	mux.Handle("GET /subscriptions/me", chain(http.HandlerFunc(subscriptionHandler.GetMySubscription), auth))
	mux.Handle("GET /subscriptions", chain(http.HandlerFunc(subscriptionHandler.GetAllSubscriptions), auth, adminOnly))
	mux.Handle("GET /subscriptions/{id}", chain(http.HandlerFunc(subscriptionHandler.GetSubscriptionByID), auth, adminOnly))
	mux.Handle("PATCH /subscriptions/{id}", chain(http.HandlerFunc(subscriptionHandler.UpdateSubscription), auth, adminOnly))

	return mux
}
