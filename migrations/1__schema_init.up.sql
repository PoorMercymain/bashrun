BEGIN;

CREATE TABLE cmd(command_id SERIAL PRIMARY KEY, command TEXT, pid INTEGER DEFAULT -2, output_text TEXT DEFAULT '', processing_status TEXT DEFAULT 'initialized', exit_status INTEGER DEFAULT NULL);

COMMIT;