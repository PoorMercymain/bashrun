package service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"

	appErrors "github.com/PoorMercymain/bashrun/errors"
	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
	"github.com/PoorMercymain/bashrun/pkg/logger"
)

var (
	_ domain.BashrunService = (*bashrunService)(nil)
)

type bashrunService struct {
	repo           domain.BashrunRepository
	sem            *semaphore.Weighted
	wg             *sync.WaitGroup
	commandContext context.Context
	sf             *singleflight.Group
}

func New(commandContext context.Context, repo domain.BashrunRepository, sem *semaphore.Weighted, wg *sync.WaitGroup) *bashrunService {
	return &bashrunService{repo: repo, sem: sem, wg: wg, commandContext: commandContext, sf: &singleflight.Group{}}
}

func (s *bashrunService) Ping(ctx context.Context) error {
	const logPrefix = "service.Ping"

	err := s.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	return nil
}

func (s *bashrunService) CreateCommand(ctx context.Context, command string) (int, error) {
	const logPrefix = "service.CreateCommand"

	id, err := s.repo.CreateCommand(ctx, command)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", logPrefix, err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		err := s.sem.Acquire(s.commandContext, 1)
		if err != nil {
			logger.Logger().Warnln("couldn't run command: semaphore didn't have enough resources")
			return
		}
		defer s.sem.Release(1)

		status, err := func() (string, error) {
			cmd := exec.CommandContext(s.commandContext, "sh", "-c", command)
			commandStdout, err := cmd.StdoutPipe()
			if err != nil {
				return "failed to create pipe", err
			}

			status, err := s.repo.ReadStatus(s.commandContext, id)
			if err != nil {
				return "failed to check status", err
			}

			if status == "stopped" {
				return "stopped", appErrors.ErrCommandStopped
			}

			err = cmd.Start()
			if err != nil {
				return "failed to start command", err
			}

			err = s.repo.UpdatePID(s.commandContext, id, cmd.Process.Pid)
			if err != nil {
				return "failed to set PID in DB", err
			}

			err = s.repo.UpdateStatus(s.commandContext, id, "started")
			if err != nil {
				return "failed to update status in DB", err
			}

			scanner := bufio.NewScanner(commandStdout)

			var outputPart string
			for scanner.Scan() {
				outputPart = scanner.Text()
				err = s.repo.UpdateOutput(s.commandContext, id, outputPart+"\n")
				if err != nil {
					return "failed to update output in DB", err
				}
			}

			var exitStatus int
			err = cmd.Wait()
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					exitStatus = exitErr.ExitCode()
				} else {
					return "failed to wait for a process to finish", err
				}
			}

			err = s.repo.UpdateExitStatus(s.commandContext, id, exitStatus)
			if err != nil {
				return "failed to update exit status in DB", err
			}

			status, err = s.repo.ReadStatus(s.commandContext, id)
			if err != nil {
				return "failed to check status", err
			}

			if status == "stopped" {
				return "stopped", appErrors.ErrCommandStopped
			}

			return "done", errors.New("")
		}()

		if err != nil {
			logger.Logger().Error(logPrefix, ": ", err.Error())
			c, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err = s.repo.UpdateStatus(c, id, status)
			if err != nil {
				logger.Logger().Error(logPrefix, ": ", err.Error())
			}
		}
	}()

	return id, nil
}

func (s *bashrunService) ListCommands(ctx context.Context, limit int, offset int) ([]domain.CommandFromDB, error) {
	const logPrefix = "service.ListCommands"

	commands, err := s.repo.ListCommands(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logPrefix, err)
	}

	return commands, nil
}

func (s *bashrunService) StopCommand(ctx context.Context, id int) error {
	const logPrefix = "service.StopCommand"

	status, err := s.repo.ReadStatus(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	if status == "created" {
		err = s.repo.UpdateStatus(ctx, id, "stopped")
		if err != nil {
			return fmt.Errorf("%s: %w", logPrefix, err)
		}

		return nil
	}

	if status != "started" {
		return appErrors.ErrCommandNotRunning
	}

	pid, err := s.repo.ReadPID(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", logPrefix, err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		_, err, _ = s.sf.Do(strconv.Itoa(id), func() (interface{}, error) {
			proc, err := os.FindProcess(pid)
			if err != nil {
				return nil, err
			}

			err = proc.Kill()
			if err != nil {
				return nil, err
			}

			err = s.repo.UpdateStatus(s.commandContext, id, "stopped")
			if err != nil {
				return nil, err
			}

			return nil, nil
		})

		if err != nil {
			logger.Logger().Error(logPrefix, ": ", err.Error())
		}
	}()

	return nil
}

func (s *bashrunService) ReadCommand(ctx context.Context, id int) (domain.CommandFromDB, error) {
	const logPrefix = "service.ReadCommand"

	command, err := s.repo.ReadCommand(ctx, id)
	if err != nil {
		return domain.CommandFromDB{}, fmt.Errorf("%s: %w", logPrefix, err)
	}

	return command, nil
}

func (s *bashrunService) ReadOutput(ctx context.Context, id int) (string, error) {
	const logPrefix = "service.ReadOutput"

	output, err := s.repo.ReadOutput(ctx, id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", logPrefix, err)
	}

	return output, nil
}
