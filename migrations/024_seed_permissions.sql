-- Semilla global de permisos. La tabla permissions es global (no por empresa),
-- por lo que se llena UNA sola vez con todos los nombres usados en el router.
INSERT INTO permissions (name, description) VALUES
    ('users:read', 'Ver usuarios y roles'),
    ('users:create', 'Crear usuarios'),
    ('users:update', 'Actualizar usuarios'),
    ('users:delete', 'Eliminar usuarios'),

    ('projects:create', 'Crear proyectos'),
    ('projects:read', 'Ver proyectos'),
    ('projects:update', 'Actualizar proyectos'),
    ('projects:delete', 'Eliminar proyectos'),

    ('clients:create', 'Crear clientes'),
    ('clients:read', 'Ver clientes'),
    ('clients:update', 'Actualizar clientes'),
    ('clients:delete', 'Eliminar clientes'),

    ('budgets:create', 'Crear presupuestos'),
    ('budgets:read', 'Ver presupuestos'),
    ('budgets:update', 'Actualizar presupuestos'),
    ('budgets:delete', 'Eliminar presupuestos'),
    ('budgets:approve', 'Aprobar presupuestos'),

    ('expenses:create', 'Registrar gastos'),
    ('expenses:read', 'Ver gastos'),
    ('expenses:update', 'Actualizar gastos'),
    ('expenses:delete', 'Eliminar gastos'),

    ('purchases:create', 'Crear órdenes de compra'),
    ('purchases:read', 'Ver órdenes de compra'),
    ('purchases:update', 'Actualizar órdenes de compra'),
    ('purchases:delete', 'Eliminar órdenes de compra'),
    ('purchases:approve', 'Aprobar órdenes de compra'),

    ('suppliers:create', 'Crear proveedores'),
    ('suppliers:read', 'Ver proveedores'),
    ('suppliers:update', 'Actualizar proveedores'),
    ('suppliers:delete', 'Eliminar proveedores'),

    ('inventory:read', 'Ver inventario'),
    ('inventory:manage', 'Gestionar inventario'),

    ('equipment:read', 'Ver equipos'),
    ('equipment:manage', 'Gestionar equipos'),
    ('equipment:assign', 'Asignar equipos'),

    ('personnel:read', 'Ver personal'),
    ('personnel:manage', 'Gestionar personal'),

    ('attendance:read', 'Ver asistencia'),
    ('attendance:mark', 'Registrar asistencia'),

    ('contractors:read', 'Ver contratistas'),
    ('contractors:manage', 'Gestionar contratistas'),
    ('contractors:pay', 'Registrar pagos a contratistas'),

    ('schedule:read', 'Ver cronograma'),
    ('schedule:update', 'Actualizar cronograma'),

    ('progress:create', 'Crear reportes de avance'),
    ('progress:read', 'Ver reportes de avance'),
    ('progress:update', 'Actualizar reportes de avance'),
    ('progress:delete', 'Eliminar reportes de avance'),

    ('photos:upload', 'Subir fotos'),
    ('photos:read', 'Ver galería de fotos'),
    ('photos:delete', 'Eliminar fotos'),

    ('invoices:create', 'Crear facturas'),
    ('invoices:read', 'Ver facturas'),
    ('invoices:update', 'Actualizar facturas'),
    ('invoices:delete', 'Eliminar facturas'),
    ('invoices:cancel', 'Anular facturas'),
    ('invoices:pay', 'Registrar pagos de facturas'),

    ('dashboard:read', 'Ver dashboard'),

    ('documents:create', 'Crear documentos'),
    ('documents:read', 'Ver documentos'),
    ('documents:update', 'Actualizar documentos'),
    ('documents:delete', 'Eliminar documentos'),

    ('notifications:read', 'Ver notificaciones'),
    ('notifications:manage', 'Gestionar notificaciones'),

    ('audits:read', 'Ver auditoría')    
ON CONFLICT (name) DO NOTHING;
