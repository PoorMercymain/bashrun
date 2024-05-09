package domain

import "context"

type BashrunService interface {
	Ping(ctx context.Context) error
	CreateCommand(ctx context.Context, command string) (int, error)
	ListCommands(ctx context.Context, limit int, offset int) ([]CommandFromDB, error)
	StopCommand(ctx context.Context, id int) error
	ReadCommand(ctx context.Context, id int) (CommandFromDB, error)
}

type BashrunRepository interface {
	Ping(ctx context.Context) error
	CreateCommand(ctx context.Context, command string) (int, error)
	UpdateOutput(ctx context.Context, id int, newOutputPart string) error
	UpdateStatus(ctx context.Context, id int, newStatus string) error
	UpdatePID(ctx context.Context, id int, pid int) error
	ListCommands(ctx context.Context, limit int, offset int) ([]CommandFromDB, error)
	UpdateExitStatus(ctx context.Context, id int, exitStatusCode int) error
	ReadStatus(ctx context.Context, id int) (string, error)
	ReadPID(ctx context.Context, id int) (int, error)
	ReadCommand(ctx context.Context, id int) (CommandFromDB, error)
}
