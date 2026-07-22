-- Evita borrar proyectos con datos relacionados.
-- Cambia todos los ON DELETE CASCADE a RESTRICT en las FK que apuntan a projects(id).

ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_project_id_fkey;
ALTER TABLE project_members ADD CONSTRAINT project_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE budgets DROP CONSTRAINT IF EXISTS budgets_project_id_fkey;
ALTER TABLE budgets ADD CONSTRAINT budgets_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE expenses DROP CONSTRAINT IF EXISTS expenses_project_id_fkey;
ALTER TABLE expenses ADD CONSTRAINT expenses_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE purchase_orders DROP CONSTRAINT IF EXISTS purchase_orders_project_id_fkey;
ALTER TABLE purchase_orders ADD CONSTRAINT purchase_orders_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE warehouses DROP CONSTRAINT IF EXISTS warehouses_project_id_fkey;
ALTER TABLE warehouses ADD CONSTRAINT warehouses_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE equipment_assignments DROP CONSTRAINT IF EXISTS equipment_assignments_project_id_fkey;
ALTER TABLE equipment_assignments ADD CONSTRAINT equipment_assignments_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE attendance DROP CONSTRAINT IF EXISTS attendance_project_id_fkey;
ALTER TABLE attendance ADD CONSTRAINT attendance_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE contractor_contracts DROP CONSTRAINT IF EXISTS contractor_contracts_project_id_fkey;
ALTER TABLE contractor_contracts ADD CONSTRAINT contractor_contracts_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_project_id_fkey;
ALTER TABLE tasks ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE daily_reports DROP CONSTRAINT IF EXISTS daily_reports_project_id_fkey;
ALTER TABLE daily_reports ADD CONSTRAINT daily_reports_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE progress_entries DROP CONSTRAINT IF EXISTS progress_entries_project_id_fkey;
ALTER TABLE progress_entries ADD CONSTRAINT progress_entries_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE project_photos DROP CONSTRAINT IF EXISTS project_photos_project_id_fkey;
ALTER TABLE project_photos ADD CONSTRAINT project_photos_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_project_id_fkey;
ALTER TABLE invoices ADD CONSTRAINT invoices_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_project_id_fkey;
ALTER TABLE payments ADD CONSTRAINT payments_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_project_id_fkey;
ALTER TABLE documents ADD CONSTRAINT documents_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_project_id_fkey;
ALTER TABLE notifications ADD CONSTRAINT notifications_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT;
