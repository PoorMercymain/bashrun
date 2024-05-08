package domain

import "context"

type BashrunService interface {
	Ping(ctx context.Context) error
}

type BashrunRepository interface {
	Ping(ctx context.Context) error
}
