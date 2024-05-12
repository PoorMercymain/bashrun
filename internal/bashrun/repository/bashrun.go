package repository

import (
	"context"
	"errors"
	"fmt"

	appErrors "github.com/PoorMercymain/bashrun/errors"
	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
	"github.com/jackc/pgx/v5"
)

var (
	_ domain.BashrunRepository = (*bashrunRepository)(nil)
)

type bashrunRepository struct {
	db *postgres
}

func New(db *postgres) *bashrunRepository {
	return &bashrunRepository{db: db}
}

func (r *bashrunRepository) Ping(ctx context.Context) error {
	const logPrefix = "repository.Ping"

	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}

func (r *bashrunRepository) CreateCommand(ctx context.Context, command string) (int, error) {
	const logPrefix = "repository.CreateCommand"

	var id int
	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, "INSERT INTO cmd(command) VALUES($1) RETURNING command_id", command).Scan(&id)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("%s: %w", logPrefix, err)
	}

	return id, nil
}

func (r *bashrunRepository) UpdateOutput(ctx context.Context, id int, newOutputPart string) error {
	const logPrefix = "repository.UpdateOutput"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "UPDATE cmd SET output_text = output_text || $1 WHERE command_id = $2", newOutputPart, id)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrRowsNotAffected
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}

func (r *bashrunRepository) UpdateStatus(ctx context.Context, id int, newStatus string) error {
	const logPrefix = "repository.UpdateStatus"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "UPDATE cmd SET processing_status = $1 WHERE command_id = $2", newStatus, id)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrRowsNotAffected
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}

func (r *bashrunRepository) UpdatePID(ctx context.Context, id int, pid int) error {
	const logPrefix = "repository.UpdatePID"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "UPDATE cmd SET pid = $1 WHERE command_id = $2", pid, id)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrRowsNotAffected
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}

func (r *bashrunRepository) UpdateExitStatus(ctx context.Context, id int, exitStatusCode int) error {
	const logPrefix = "repository.UpdateExitStatus"

	err := r.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "UPDATE cmd SET exit_status = $1 WHERE command_id = $2", exitStatusCode, id)
		if err != nil {
			return err
		}

		if tag.RowsAffected() == 0 {
			return appErrors.ErrRowsNotAffected
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}

func (r *bashrunRepository) ListCommands(ctx context.Context, limit int, offset int) ([]domain.CommandFromDB, error) {
	const logPrefix = "repository.ListCommands"

	rows, err := r.db.Query(ctx, "SELECT command_id, command, pid, output_text, processing_status, exit_status FROM cmd ORDER BY command_id ASC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logPrefix, err)
	}
	defer rows.Close()

	commands := make([]domain.CommandFromDB, 0)
	for rows.Next() {
		var command domain.CommandFromDB

		err = rows.Scan(&command.ID, &command.Command, &command.PID, &command.Output, &command.Status, &command.ExitStatus)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logPrefix, err)
		}

		commands = append(commands, command)
	}

	if len(commands) == 0 {
		return nil, appErrors.ErrNoRows
	}

	return commands, nil
}

func (r *bashrunRepository) ReadStatus(ctx context.Context, id int) (string, error) {
	const logPrefix = "repository.ReadStatus"

	var status string
	err := r.db.QueryRow(ctx, "SELECT processing_status FROM cmd WHERE command_id = $1", id).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", appErrors.ErrNoRows
		}

		return "", fmt.Errorf("%s: %w", logPrefix, err)
	}

	return status, nil
}

func (r *bashrunRepository) ReadPID(ctx context.Context, id int) (int, error) {
	const logPrefix = "repository.ReadPID"

	var pid int
	err := r.db.QueryRow(ctx, "SELECT pid FROM cmd WHERE command_id = $1", id).Scan(&pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, appErrors.ErrNoRows
		}

		return 0, fmt.Errorf("%s: %w", logPrefix, err)
	}

	return pid, nil
}

func (r *bashrunRepository) ReadCommand(ctx context.Context, id int) (domain.CommandFromDB, error) {
	const logPrefix = "repository.ReadCommand"

	var command domain.CommandFromDB
	err := r.db.QueryRow(ctx, "SELECT command_id, command, pid, output_text, processing_status, exit_status FROM cmd WHERE command_id = $1", id).Scan(&command.ID, &command.Command, &command.PID, &command.Output, &command.Status, &command.ExitStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CommandFromDB{}, appErrors.ErrNoRows
		}

		return domain.CommandFromDB{}, fmt.Errorf("%s: %w", logPrefix, err)
	}

	return command, nil
}

func (r *bashrunRepository) ReadOutput(ctx context.Context, id int) (string, error) {
	const logPrefix = "repository.ReadOutput"

	var output string
	err := r.db.QueryRow(ctx, "SELECT output_text FROM cmd WHERE command_id = $1", id).Scan(&output)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", appErrors.ErrNoRows
		}

		return "", fmt.Errorf("%s: %w", logPrefix, err)
	}

	return output, nil
}
