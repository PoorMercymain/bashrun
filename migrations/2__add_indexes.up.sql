BEGIN;

-- тут не факт, что нужен индекс, зависит от того, что будет чаще производиться - чтение или вставка/обновление
CREATE INDEX IF NOT EXISTS idx_cmd ON cmd(command_id, command, pid, output_text, processing_status, exit_status);

COMMIT;