# Constructora ERP — Backend

API REST escrita en Go para la gestión integral de empresas constructoras. El sistema está diseñado bajo un esquema de **multi-tenancy** donde cada empresa (constructora) opera sus datos de forma aislada a través de una relación `company_id` que atraviesa todas las entidades del sistema.

## Arquitectura

El proyecto sigue una arquitectura limpia en capas dentro del paquete `internal/`:

- **Handler** — capa de transporte, recibe y responde peticiones HTTP.
- **Service** — lógica de negocio y validaciones.
- **Repository** — acceso a base de datos (PostgreSQL).

Cada módulo del dominio de construcción vive en su propio subpaquete dentro de `internal/`, manteniendo independencia y cohesión.

## Módulos del sistema

| Módulo           | Descripción |
|-----------------|-------------|
| **Users**       | Registro y autenticación de usuarios por empresa, con un sistema de roles y permisos (RBAC) que relaciona usuarios ↔ roles ↔ permisos mediante tablas intermedias. |
| **Projects**    | Gestión de proyectos de construcción, entidad central que agrupa el resto de módulos. |
| **Clients**     | Administración de clientes asociados a cada proyecto. |
| **Budgets**     | Elaboración y seguimiento de presupuestos por proyecto. |
| **Expenses**    | Registro y control de gastos vinculados a proyectos. |
| **Purchase Orders** | Órdenes de compra para insumos y materiales. |
| **Suppliers**   | Catálogo y gestión de proveedores. |
| **Inventory**   | Control de materiales, almacenes (bodegas) y movimientos de stock. |
| **Equipment**   | Gestión de maquinaria y equipo, incluyendo asignaciones y programación de mantenimientos. |
| **Personnel**   | Administración de posiciones/cargos, empleados y contratos laborales. |
| **Attendance**  | Control de asistencia del personal en obra. |
| **Contractors** | Gestión de contratistas externos, sus contratos y pagos. |
| **Schedule**    | Planificación de tareas, hitos (milestones) y dependencias entre actividades (diagrama de Gantt). |
| **Progress**    | Reportes diarios de avance de obra. |
| **Photos**      | Registro de metadatos de evidencia fotográfica por proyecto. |
| **Payments / Invoices** | Facturación, pagos y cancelación de facturas. |
| **Financial Dashboard** | Resumen financiero consolidado por proyecto. |
| **Documents**   | Gestión de tipos documentales, documentos y control de versiones. |
| **Notifications** | Sistema de notificaciones internas por usuario. |
| **Audit Logs**  | Traza de auditoría para rastrear cambios y accesos. |
| **Subscriptions** | Planes de suscripción por empresa con control de límites (proyectos, usuarios, almacenamiento) y facturación recurrente. |

## Seguridad

- Autenticación mediante **JWT** con claims de `user_id` y `company_id`.
- Middleware de autorización que valida el token en cada petición protegida.
- Middleware de verificación de suscripción activa que controla el acceso según el plan contratado.
- Aislamiento multitenant: todas las consultas están filtradas por `company_id`, asegurando que cada empresa solo acceda a sus propios datos.

## Base de datos: **PostgreSQL** 

## Stack técnico

- **Lenguaje:** Go
- **Base de datos:** PostgreSQL
- **Autenticación:** JWT (golang-jwt)
- **Enrutamiento:** `net/http.ServeMux` (Go 1.22+, con path variables)
- **Migraciones:** Archivos SQL secuenciales ejecutados manualmente o mediante herramienta externa
