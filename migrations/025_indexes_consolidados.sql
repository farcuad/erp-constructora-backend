-- ============================================================================
-- ÍNDICES CONSOLIDADOS — TODOS LOS MÓDULOS DEL ERP
-- ============================================================================
-- Reemplaza a los scripts 025_indexes_projects.sql .. 044_indexes_subscriptions.sql.
-- Cada bloque está agrupado por módulo; los índices referencian su módulo.
--
-- LEE ESTO ANTES DE CORRER CUALQUIER COSA
-- ----------------------------------------------------------------------------
-- 1) PostgreSQL crea índices automáticamente para PRIMARY KEY y para UNIQUE.
--    NO los crea para FOREIGN KEY. Ese es el hueco que tapan estos scripts.
--
-- 2) Usamos CREATE INDEX normal (no CONCURRENTLY) porque toma un lock corto
--    que en tablas pequeñas es instantáneo, y porque CONCURRENTLY NO puede
--    correr dentro de una transacción (el editor SQL de Supabase y `psql -1`
--    envuelven todo en una, y te va a fallar).
--    Si alguna tabla ya tiene cientos de miles de filas y no puedes bloquear
--    escrituras, corre ESA sentencia sola, sin transacción, así:
--        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_... ON ...;
--
-- 3) Todo es IF NOT EXISTS / IF EXISTS: puedes correr el script las veces que
--    quieras, no rompe nada.
--
-- 4) Al final de cada módulo hay un ANALYZE. Sin él, el planner sigue usando
--    estadísticas viejas y puede ignorar los índices que acabas de crear.
--
-- 5) Las queries de VERIFICACIÓN del 045_indexes_verificacion.sql están al
--    final, listas para ejecutar a mano unos días después.
-- ============================================================================


-- ============================================================================
-- MÓDULO: PROJECTS  (el más caliente de todo el ERP)
-- ============================================================================
-- Sirve a:
--   GET /projects                    -> projects.GetAll
--   subscriptions.CountActiveProjects (se dispara al crear proyectos)
--   Todas las FK ...project_id que quedaron en ON DELETE RESTRICT (mig. 022)

-- [CRÍTICO] GET /projects
--   Query: WHERE company_id = $1 ORDER BY created_at DESC
--   Las 2 columnas del índice cubren el filtro Y el orden: Postgres hace un
--   solo seek y lee las filas ya ordenadas, sin nodo de Sort.
--   También resuelve `SELECT COUNT(*) FROM projects WHERE company_id = $1`
--   (CountActiveProjects) con un Index Only Scan, sin tocar la tabla.
CREATE INDEX IF NOT EXISTS idx_projects_company_created
    ON projects (company_id, created_at DESC);

-- [FK] project_members.user_id
--   users -> project_members es ON DELETE CASCADE. Sin este índice, borrar un
--   usuario recorre project_members entera. (El lado project_id ya lo cubre
--   la PRIMARY KEY (project_id, user_id)).
CREATE INDEX IF NOT EXISTS idx_project_members_user
    ON project_members (user_id);

ANALYZE projects;


-- ============================================================================
-- MÓDULO: USERS / ROLES
-- ============================================================================
-- Sirve a:
--   POST /login                -> users.GetEmailUser / GetEmailUserWithDetails
--   POST /register             -> users.EmailExists
--   GET  /users                -> users.GetUsersExcludingAdmin
--   GET  /roles                -> users.GetRolesByCompanyID
--   subscriptions.CountActiveUsers

-- [CRÍTICO] Login. Query: WHERE email = $1
--   OJO: la constraint UNIQUE (company_id, email) NO sirve aquí, porque en un
--   índice compuesto solo puedes hacer seek si filtras por la PRIMERA columna.
--   Filtrando solo por email, Postgres hace Seq Scan de toda la tabla users.
--   No puede ser UNIQUE: el email es único POR EMPRESA, no globalmente.
CREATE INDEX IF NOT EXISTS idx_users_email
    ON users (email);

-- [ALTO] CountActiveUsers: WHERE company_id = $1 AND is_active = true
--   Índice PARCIAL: solo indexa las filas activas, así que ocupa menos y se
--   mantiene más barato. También cubre el WHERE u.company_id = $1 de GET /users.
CREATE INDEX IF NOT EXISTS idx_users_company_active
    ON users (company_id)
    WHERE is_active;

-- [FK] user_roles.role_id
--   roles -> user_roles es ON DELETE CASCADE. Sin índice, borrar un rol recorre
--   user_roles entera. (El lado user_id lo cubre la PK (user_id, role_id), que
--   además es la que usa el JOIN de GetUsersExcludingAdmin).
CREATE INDEX IF NOT EXISTS idx_user_roles_role
    ON user_roles (role_id);

ANALYZE users;
ANALYZE user_roles;


-- ============================================================================
-- MÓDULO: CLIENTS
-- ============================================================================
-- Sirve a:
--   GET /clients -> clients.GetByCompany

-- [FK] client_contacts.client_id
--   clients -> client_contacts es ON DELETE CASCADE. Sin índice, borrar un
--   cliente recorre client_contacts entera.
CREATE INDEX IF NOT EXISTS idx_client_contacts_client
    ON client_contacts (client_id);

ANALYZE client_contacts;


-- ============================================================================
-- MÓDULO: BUDGETS
-- ============================================================================
-- Sirve a:
--   GET /budgets/{project_id} -> budgets.GetBudgetsProjectID
--   GET /dashboard/financial/{project_id} -> SUM(total_amount) FROM budgets

-- [FK] budget_items.budget_version_id
--   budget_versions -> budget_items es ON DELETE CASCADE, sin índice.
CREATE INDEX IF NOT EXISTS idx_budget_items_version
    ON budget_items (budget_version_id);

-- [FK] budget_versions.changed_by
--   users -> budget_versions es ON DELETE SET NULL. Sin índice, borrar un
--   usuario recorre budget_versions entera.
CREATE INDEX IF NOT EXISTS idx_budget_versions_changed_by
    ON budget_versions (changed_by);

ANALYZE budgets;
ANALYZE budget_items;


-- ============================================================================
-- MÓDULO: EXPENSES
-- ============================================================================
-- Sirve a:
--   GET /expenses/{project_id} -> expense.GetByProject
--   GET /dashboard/financial/{project_id} -> SUM(amount) FROM expenses

-- [CRÍTICO] GET /expenses/{project_id}
--   Query: WHERE company_id=$1 AND project_id=$2 ORDER BY expense_date DESC
--   La tabla expenses hoy solo tiene el índice de la PRIMARY KEY: cualquier
--   consulta por proyecto es un Seq Scan completo.
--   Las 3 columnas cubren filtro + orden, y el INCLUDE (amount) hace que el
--   SUM(amount) del dashboard se resuelva sin tocar el heap.
CREATE INDEX IF NOT EXISTS idx_expenses_company_project_date
    ON expenses (company_id, project_id, expense_date DESC)
    INCLUDE (amount);

-- [FK] expense_documents.expense_id
--   expenses -> expense_documents es ON DELETE CASCADE, sin índice.
CREATE INDEX IF NOT EXISTS idx_expense_documents_expense
    ON expense_documents (expense_id);

-- [FK] expenses.user_id
--   La FK a users no declara acción (= NO ACTION), pero Postgres igual valida
--   al borrar el usuario, y sin índice ese chequeo recorre expenses entera.
CREATE INDEX IF NOT EXISTS idx_expenses_user
    ON expenses (user_id);

ANALYZE expenses;
ANALYZE expense_documents;


-- ============================================================================
-- MÓDULO: PURCHASE ORDERS
-- ============================================================================
-- Sirve a:
--   GET /purcharse/{project_id} -> purchase.GetOrdersByProject
--   GET /dashboard/financial/{project_id} -> SUM(total_amount) FROM purchase_orders

-- [CRÍTICO] Un solo índice cubre las dos consultas:
--   a) GetOrdersByProject : WHERE project_id = $1
--      -> usa la primera columna, seek directo.
--   b) Dashboard          : WHERE project_id = $2 AND company_id = $1
--      -> usa las dos columnas + lee total_amount desde el propio índice.
--   Por eso project_id va PRIMERO: es la columna que ambas comparten.
CREATE INDEX IF NOT EXISTS idx_purchase_orders_project_company
    ON purchase_orders (project_id, company_id)
    INCLUDE (total_amount);

-- [FK] purchase_order_items.purchase_order_id
--   purchase_orders -> purchase_order_items es ON DELETE CASCADE, sin índice.
CREATE INDEX IF NOT EXISTS idx_purchase_order_items_order
    ON purchase_order_items (purchase_order_id);

-- [FK] purchase_orders.supplier_id
--   Es ON DELETE RESTRICT: cada DELETE /supplier/{id} tiene que probar que no
--   haya órdenes colgando. Sin índice, ese chequeo recorre purchase_orders.
CREATE INDEX IF NOT EXISTS idx_purchase_orders_supplier
    ON purchase_orders (supplier_id);

-- [FK] purchase_orders.user_id
CREATE INDEX IF NOT EXISTS idx_purchase_orders_user
    ON purchase_orders (user_id);

ANALYZE purchase_orders;
ANALYZE purchase_order_items;


-- ============================================================================
-- MÓDULO: SUPPLIERS
-- ============================================================================
-- Sirve a:
--   GET /supplier -> suppliers.GetSuppliersByCompany

-- [FK] supplier_contacts.supplier_id
--   suppliers -> supplier_contacts es ON DELETE CASCADE, sin índice.
CREATE INDEX IF NOT EXISTS idx_supplier_contacts_supplier
    ON supplier_contacts (supplier_id);

ANALYZE supplier_contacts;


-- ============================================================================
-- MÓDULO: INVENTORY (materiales, almacenes, stock)
-- ============================================================================
-- Sirve a:
--   GET /materials                     -> inventory.GetMaterials
--   GET /warehouses                    -> inventory.GetWarehouses
--   GET /inventory/stock/{warehouse_id}-> inventory.GetStockByWarehouse

-- [CRÍTICO] GET /warehouses -> WHERE company_id = $1
--   warehouses solo tiene UNIQUE (project_id, name): su primera columna es
--   project_id, así que filtrar por company_id no puede aprovecharla y termina
--   en Seq Scan. Este índice tapa ese hueco.
CREATE INDEX IF NOT EXISTS idx_warehouses_company
    ON warehouses (company_id);

-- [FK] materials.category_id -> material_categories ON DELETE SET NULL
CREATE INDEX IF NOT EXISTS idx_materials_category
    ON materials (category_id);

-- [FK + futuro] stock_movements: es la tabla que más crece del módulo (una
--   fila por cada entrada/salida) y hoy no tiene ningún índice fuera de la PK.
--   Las 3 FK (warehouse_id, material_id, user_id) son CASCADE sin índice.
CREATE INDEX IF NOT EXISTS idx_stock_movements_warehouse_created
    ON stock_movements (warehouse_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_stock_movements_material
    ON stock_movements (material_id);

CREATE INDEX IF NOT EXISTS idx_stock_movements_user
    ON stock_movements (user_id);

ANALYZE warehouses;
ANALYZE materials;
ANALYZE stock_movements;


-- ============================================================================
-- MÓDULO: EQUIPMENT (maquinaria, asignaciones, mantenimientos)
-- ============================================================================
-- Sirve a:
--   GET /equipment                              -> GetEquipmentByCompany
--   GET /equipment/types                        -> GetEquipmentType
--   GET /equipment/assignments/{equipment_id}   -> GetEquipementAssignment
--   GET /equipment/maintenances/{equipment_id}  -> GetEquipementMaintenance

-- [CRÍTICO] GET /equipment/assignments/{equipment_id}
--   Query: WHERE equipment_id = $1
--   equipment_assignments no tiene NINGÚN índice fuera de la PK -> Seq Scan.
--   Agrego start_date DESC para que la asignación más reciente salga primero
--   sin costo extra (es la que casi siempre interesa).
CREATE INDEX IF NOT EXISTS idx_equipment_assignments_equipment
    ON equipment_assignments (equipment_id, start_date DESC);

-- [CRÍTICO] GET /equipment/maintenances/{equipment_id}
--   Query: WHERE equipment_id = $1. Mismo caso: hoy es Seq Scan.
CREATE INDEX IF NOT EXISTS idx_maintenance_records_equipment
    ON maintenance_records (equipment_id, maintenance_date DESC);

-- [FK] equipment_assignments.project_id
--   Quedó en ON DELETE RESTRICT (migración 022): cada DELETE /projects/{id}
--   valida esta tabla. Sin índice, es un Seq Scan por cada borrado.
CREATE INDEX IF NOT EXISTS idx_equipment_assignments_project
    ON equipment_assignments (project_id);

-- [FK] equipment.type_id
--   La migración 010 lo pasó a ON DELETE RESTRICT: cada
--   DELETE /equipment/types/{id} recorre equipment entera sin este índice.
CREATE INDEX IF NOT EXISTS idx_equipment_type
    ON equipment (type_id);

-- [FK] equipment_assignments.assigned_by -> users
CREATE INDEX IF NOT EXISTS idx_equipment_assignments_assigned_by
    ON equipment_assignments (assigned_by);

ANALYZE equipment_assignments;
ANALYZE maintenance_records;
ANALYZE equipment;


-- ============================================================================
-- MÓDULO: PERSONNEL (empleados, cargos, contratos laborales)
-- ============================================================================
-- Sirve a:
--   GET /employees            -> personnel.GetEmployeesByCompany
--   GET /positions            -> personnel.GetPosition
--   GET /contracts/{project_id} -> personnel.GetContract

-- [CRÍTICO] GET /contracts/{project_id}
--   Query: WHERE project_id = $1
--   contracts no tiene índice en project_id -> Seq Scan de todos los contratos
--   de todas las empresas. Además la FK quedó en ON DELETE SET NULL, así que
--   borrar un proyecto también recorre la tabla entera.
CREATE INDEX IF NOT EXISTS idx_contracts_project
    ON contracts (project_id);

-- [FK] contracts.employee_id -> employees ON DELETE CASCADE
CREATE INDEX IF NOT EXISTS idx_contracts_employee
    ON contracts (employee_id);

-- [FK] employees.position_id -> positions ON DELETE SET NULL
--   Cada DELETE /positions/{id} valida esta columna.
CREATE INDEX IF NOT EXISTS idx_employees_position
    ON employees (position_id);

ANALYZE contracts;
ANALYZE employees;


-- ============================================================================
-- MÓDULO: ATTENDANCE
-- ============================================================================
-- Sirve a:
--   GET /attendance/{project_id} -> attendance.GetAttendanceByProjectAndDate

-- [FK] attendance_logs.employee_id -> employees ON DELETE CASCADE
--   Es la única FK del módulo sin índice: borrar un empleado recorre
--   attendance_logs entera (la tabla que más crece del módulo: una fila por
--   empleado por día trabajado).
CREATE INDEX IF NOT EXISTS idx_attendance_logs_employee
    ON attendance_logs (employee_id);

ANALYZE attendance_logs;


-- ============================================================================
-- MÓDULO: CONTRACTORS (subcontratistas, contratos, pagos)
-- ============================================================================
-- Sirve a:
--   GET /contractors                        -> contractors.GetContracts
--   GET /contractors/contracts/{project_id} -> contractors.GetContractsByProject
--   GET /contractors/payments               -> contractors.GetContractPayment

-- [CRÍTICO] GET /contractors/contracts/{project_id}
--   Query: WHERE project_id = $1. Hoy es Seq Scan.
--   Además la FK quedó en ON DELETE RESTRICT (migración 022), así que este
--   índice también acelera cada DELETE /projects/{id}.
CREATE INDEX IF NOT EXISTS idx_contractor_contracts_project
    ON contractor_contracts (project_id);

-- [FK] contractor_contracts.contractor_id -> ON DELETE RESTRICT
--   Cada DELETE /contractors/{id} valida esta columna.
CREATE INDEX IF NOT EXISTS idx_contractor_contracts_contractor
    ON contractor_contracts (contractor_id);

-- [ALTO] contractor_payments.contract_id
--   Es FK CASCADE sin índice, y es la columna por la que DEBERÍA filtrar
--   GET /contractors/payments (ver la nota de abajo).
CREATE INDEX IF NOT EXISTS idx_contractor_payments_contract
    ON contractor_payments (contract_id, payment_date DESC);

-- [FK] contractor_payments.user_id -> users
CREATE INDEX IF NOT EXISTS idx_contractor_payments_user
    ON contractor_payments (user_id);

ANALYZE contractor_contracts;
ANALYZE contractor_payments;


-- ============================================================================
-- MÓDULO: SCHEDULE (cronograma / tareas)
-- ============================================================================
-- Sirve a:
--   GET /schedule/{project_id} -> schedule.GetByProject

-- [ALTO] GET /schedule/{project_id}
--   Query: WHERE project_id = $1 ORDER BY created_at ASC
--   Ya existe idx_tasks_project (project_id) desde la migración 014, pero solo
--   cubre el filtro: Postgres tiene que ordenar el resultado aparte. Con
--   created_at dentro del índice, el ORDER BY sale gratis.
CREATE INDEX IF NOT EXISTS idx_tasks_project_created
    ON tasks (project_id, created_at);

-- El índice viejo queda contenido en el nuevo (mismo prefijo project_id), así
-- que mantenerlo solo cuesta escrituras. Se elimina.
DROP INDEX IF EXISTS idx_tasks_project;

ANALYZE tasks;


-- ============================================================================
-- MÓDULO: PROGRESS (reportes diarios y avance de obra)
-- ============================================================================
-- Sirve a:
--   GET /progress/{project_id} -> progress.GetReportWithProgress

-- [ALTO] GET /progress/{project_id}
--   Query: WHERE dr.company_id=$1 AND dr.project_id=$2 AND dr.report_date=$3
--   El índice de la migración 015 solo llega hasta project_id; sumando
--   report_date las tres columnas son igualdades y el seek devuelve UNA fila
--   directa, sin filtrar nada después.
CREATE INDEX IF NOT EXISTS idx_daily_reports_company_project_date
    ON daily_reports (company_id, project_id, report_date);

-- El viejo es un prefijo exacto del nuevo: redundante.
DROP INDEX IF EXISTS idx_daily_reports_company_project;

-- [FK] progress_entries.task_id -> tasks ON DELETE CASCADE
--   Sin índice, cada DELETE /schedule/tasks/{id} recorre progress_entries.
CREATE INDEX IF NOT EXISTS idx_progress_entries_task
    ON progress_entries (task_id);

-- [FK] progress_entries.project_id -> ON DELETE RESTRICT (migración 022)
CREATE INDEX IF NOT EXISTS idx_progress_entries_project
    ON progress_entries (project_id);

-- [FK] daily_reports.user_id -> users
CREATE INDEX IF NOT EXISTS idx_daily_reports_user
    ON daily_reports (user_id);

ANALYZE daily_reports;
ANALYZE progress_entries;


-- ============================================================================
-- MÓDULO: PHOTOS (evidencia fotográfica)
-- ============================================================================
-- Sirve a:
--   GET /photos/{project_id} -> photos.GetByProject

-- [ALTO] GET /photos/{project_id}
--   Query: WHERE company_id=$1 AND project_id=$2 ORDER BY created_at DESC
--   El índice de la migración 016 cubre el filtro pero deja el ORDER BY a un
--   nodo de Sort. Es una galería que crece rápido (una fila por foto), así que
--   conviene que el orden venga del índice.
CREATE INDEX IF NOT EXISTS idx_photos_company_project_created
    ON project_photos (company_id, project_id, created_at DESC);

-- El viejo es prefijo exacto del nuevo: redundante.
DROP INDEX IF EXISTS idx_photos_company_project;

-- [FK] project_photos.user_id -> users
CREATE INDEX IF NOT EXISTS idx_photos_user
    ON project_photos (user_id);

ANALYZE project_photos;


-- ============================================================================
-- MÓDULO: PAYMENTS / INVOICES (facturación)
-- ============================================================================
-- Sirve a:
--   GET /invoices/{id}                    -> payments.GetByID
--   GET /invoices/project/{project_id}    -> payments.GetByProject
--   GET /invoices/payments/{invoice_id}   -> payments.GetPaymentsByInvoice
--   GET /dashboard/financial/{project_id} -> 4 sub-SELECT sobre invoices/payments
--   GET /subscriptions/{id}               -> subscriptions.GetPaymentsByCompany

-- [ALTO] Dashboard financiero. Las cuatro sub-consultas filtran
--     WHERE company_id=$1 AND project_id=$2 AND type='EMITTED'|'RECEIVED'
--   El idx_invoices_lookup de la migración 017 ya resuelve eso perfecto
--   (3 igualdades sobre las 3 primeras columnas). Lo único que le falta es
--   traer total_amount sin ir al heap: por eso se recrea con INCLUDE.
--   También cubre GET /invoices/project/{project_id}, que filtra por las dos
--   primeras columnas.
--   Se crea con nombre nuevo y recién después se borra el viejo, para que la
--   tabla nunca quede un instante sin índice.
CREATE INDEX IF NOT EXISTS idx_invoices_company_project_type_status
    ON invoices (company_id, project_id, type, status)
    INCLUDE (total_amount);

DROP INDEX IF EXISTS idx_invoices_lookup;

-- [CRÍTICO] subscriptions.GetPaymentsByCompany
--   Query: WHERE p.company_id = $1 ORDER BY p.payment_date DESC
--   payments solo tenía idx_payments_invoice (invoice_id): filtrar por
--   company_id es Seq Scan sobre la tabla de pagos completa.
CREATE INDEX IF NOT EXISTS idx_payments_company_date
    ON payments (company_id, payment_date DESC);

-- [FK] payments.project_id -> ON DELETE RESTRICT (migración 022)
CREATE INDEX IF NOT EXISTS idx_payments_project
    ON payments (project_id);

-- [FK] invoices.client_id / supplier_id / contractor_id -> ON DELETE SET NULL
--   Cada DELETE de cliente, proveedor o contratista valida estas 3 columnas.
--   Van como índices PARCIALES: por diseño una factura llena SOLO una de las
--   tres (EMITTED usa client_id, RECEIVED usa supplier_id o contractor_id), así
--   que ~2/3 de las filas son NULL en cada columna y no vale la pena indexarlas.
CREATE INDEX IF NOT EXISTS idx_invoices_client
    ON invoices (client_id) WHERE client_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_invoices_supplier
    ON invoices (supplier_id) WHERE supplier_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_invoices_contractor
    ON invoices (contractor_id) WHERE contractor_id IS NOT NULL;

ANALYZE invoices;
ANALYZE payments;


-- ============================================================================
-- MÓDULO: DOCUMENTS (gestión documental y versiones)
-- ============================================================================
-- Sirve a:
--   GET /documents/types                  -> documents.GetTypes
--   GET /documents/project/{project_id}   -> documents.GetByProject
--   GET /documents/{id}                   -> documents.GetByID
--   GET /documents/versions/{document_id} -> documents.GetVersions

-- [ALTO] GET /documents/project/{project_id}
--   Query: WHERE company_id=$1 AND project_id=$2 ORDER BY created_at DESC
--   El idx_documents_lookup de la migración 018 termina en document_type_id,
--   columna por la que NINGUNA query filtra. Cambiándola por created_at, el
--   índice pasa a cubrir también el ORDER BY.
CREATE INDEX IF NOT EXISTS idx_documents_company_project_created
    ON documents (company_id, project_id, created_at DESC);

-- [FK] documents.document_type_id -> ON DELETE RESTRICT
--   Se separa en su propio índice: cada DELETE /documents/types/{id} valida
--   esta columna, y ya no la cubre el índice de arriba.
CREATE INDEX IF NOT EXISTS idx_documents_type
    ON documents (document_type_id);

-- Recién ahora se borra el viejo, con los dos reemplazos ya en su lugar.
DROP INDEX IF EXISTS idx_documents_lookup;

-- [REDUNDANTE] idx_document_versions (document_id, version_number) de la
--   migración 018 es una copia exacta del índice que Postgres ya crea solo
--   para UNIQUE (document_id, version_number). Duplicado 1:1: ocupa el doble
--   de disco y frena cada subida de versión sin aportar nada.
--   Ese índice único es además el que resuelve GET /documents/versions/{id}
--   con su ORDER BY version_number DESC.
DROP INDEX IF EXISTS idx_document_versions;

-- [FK] document_versions.user_id -> users
CREATE INDEX IF NOT EXISTS idx_document_versions_user
    ON document_versions (user_id);

ANALYZE documents;
ANALYZE document_versions;


-- ============================================================================
-- MÓDULO: NOTIFICATIONS
-- ============================================================================
-- Sirve a:
--   GET /notifications -> notifications.GetUserNotifications

-- [REDUNDANTE] idx_notification_reads_user (user_id, is_read) queda tapado por
--   idx_notification_reads_user_unread (company_id, user_id, is_read): la
--   única query del módulo filtra SIEMPRE por company_id + user_id, así que
--   usa el compuesto y nunca el de dos columnas. Sobra.
DROP INDEX IF EXISTS idx_notification_reads_user;

-- [FK] notifications.project_id -> ON DELETE RESTRICT (migración 022)
--   Parcial porque la columna es opcional (las alertas globales de empresa
--   van con project_id NULL) y esas filas no aportan nada al índice.
CREATE INDEX IF NOT EXISTS idx_notifications_project
    ON notifications (project_id) WHERE project_id IS NOT NULL;

-- [MEDIO] El ORDER BY n.created_at DESC de GetUserNotifications ordena sobre
--   notifications. Con company_id delante, el índice sirve tanto para ese
--   orden como para el filtro por empresa.
CREATE INDEX IF NOT EXISTS idx_notifications_company_created
    ON notifications (company_id, created_at DESC);

-- Sustituye a idx_notifications_company (company_id), que es su prefijo.
DROP INDEX IF EXISTS idx_notifications_company;

ANALYZE notifications;
ANALYZE notification_reads;


-- ============================================================================
-- MÓDULO: AUDIT LOGS
-- ============================================================================
-- Sirve a:
--   GET /audits-logs -> audit.GetByCompany

-- [CRÍTICO] GET /audits-logs
--   Query: WHERE company_id = $1 ORDER BY created_at DESC LIMIT 100
--   El idx_audit_logs_company_action (company_id, action) de la migración 020
--   encuentra las filas, pero como la segunda columna es `action` y no
--   `created_at`, Postgres tiene que traer TODOS los logs de la empresa,
--   ordenarlos en memoria y recién ahí aplicar el LIMIT 100.
--   Con created_at en el índice, lee las 100 primeras y para. La diferencia
--   crece sola: audit_logs es la tabla que más rápido crece de todo el ERP.
CREATE INDEX IF NOT EXISTS idx_audit_logs_company_created
    ON audit_logs (company_id, created_at DESC);

-- [FK] audit_logs.user_id -> users
CREATE INDEX IF NOT EXISTS idx_audit_logs_user
    ON audit_logs (user_id);

ANALYZE audit_logs;


-- ============================================================================
-- MÓDULO: SUBSCRIPTIONS
-- ============================================================================
-- Sirve a:
--   GET /subscriptions/me   -> subscriptions.GetByCompany
--   GET /subscriptions      -> subscriptions.GetAllWithCompany  (superadmin)
--   GET /subscriptions/{id} -> subscriptions.GetByID            (superadmin)
--   middlewares.RequireActiveSubscription -> GetByCompany, en CADA request
--                                            protegido del ERP

-- [REDUNDANTE] idx_subscriptions_company (company_id) es una copia exacta del
--   índice que Postgres crea solo para UNIQUE (company_id). Duplicado 1:1.
--   Y es justo la tabla que más se lee del sistema, así que cada escritura
--   estaba manteniendo dos estructuras idénticas.
DROP INDEX IF EXISTS idx_subscriptions_company;

ANALYZE companies_subscriptions;


-- ============================================================================
-- VERIFICACIÓN — no crea nada, son consultas para ejecutar a mano
-- ============================================================================
-- Córrelas DESPUÉS de aplicar este script y otra vez unos días más tarde, con
-- tráfico real encima.
-- ============================================================================


-- ----------------------------------------------------------------------------
-- 1) ¿Se está usando el índice de /projects?
--    Busca "Index Scan using idx_projects_company_created". Si dice
--    "Seq Scan", el índice no existe o la tabla es tan chica que Postgres
--    prefiere leerla entera (con <100 filas eso es normal y correcto).
-- ----------------------------------------------------------------------------
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, company_id, name, client_name, location, start_date, end_date,
       budget, status_id, created_at, updated_at
FROM projects
WHERE company_id = '00000000-0000-0000-0000-000000000000'  -- <- tu company_id
ORDER BY created_at DESC;


-- ----------------------------------------------------------------------------
-- 2) Índices que NADIE usa (candidatos a borrar).
--    idx_scan = 0 después de días de tráfico = índice muerto: solo frena
--    los INSERT/UPDATE. Ignora los que terminan en _pkey o _key.
-- ----------------------------------------------------------------------------
SELECT schemaname,
       relname   AS tabla,
       indexrelname AS indice,
       idx_scan  AS veces_usado,
       pg_size_pretty(pg_relation_size(indexrelid)) AS peso
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan ASC, pg_relation_size(indexrelid) DESC;


-- ----------------------------------------------------------------------------
-- 3) Foreign keys que TODAVÍA no tienen índice.
--    Debería salir vacío (o casi) después de correr este script. Cada fila que
--    aparezca es un DELETE que hace Seq Scan de la tabla hija.
-- ----------------------------------------------------------------------------
SELECT c.conrelid::regclass AS tabla,
       a.attname            AS columna_fk,
       c.conname            AS constraint
FROM pg_constraint c
JOIN pg_attribute a
  ON a.attrelid = c.conrelid
 AND a.attnum   = c.conkey[1]          -- primera columna de la FK
WHERE c.contype = 'f'
  AND c.connamespace = 'public'::regnamespace
  AND NOT EXISTS (
      -- ¿existe algún índice cuya PRIMERA columna sea esa misma?
      -- (indkey es un int2vector: su primer elemento es el subíndice 0)
      SELECT 1
      FROM pg_index i
      WHERE i.indrelid = c.conrelid
        AND (i.indkey::smallint[])[0] = c.conkey[1]
  )
ORDER BY 1, 2;


-- ----------------------------------------------------------------------------
-- 4) Tablas que siguen leyéndose enteras (Seq Scan) pese a los índices.
--    seq_scan alto + seq_tup_read alto en una tabla grande = falta un índice
--    o alguna query no filtra por lo que debería.
-- ----------------------------------------------------------------------------
SELECT relname AS tabla,
       seq_scan,
       seq_tup_read,
       idx_scan,
       n_live_tup AS filas_aprox
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY seq_tup_read DESC
LIMIT 20;


-- ----------------------------------------------------------------------------
-- 5) Índices marcados como INVALID.
--    Solo puede pasar si usaste CREATE INDEX CONCURRENTLY y falló a medias.
--    Un índice invalid no se usa pero sí se mantiene: hay que borrarlo y
--    volver a crearlo.
-- ----------------------------------------------------------------------------
SELECT indexrelid::regclass AS indice_invalido
FROM pg_index
WHERE NOT indisvalid;