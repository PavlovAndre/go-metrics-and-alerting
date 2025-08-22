package retry

import (
	"context"
	"errors"
	"fmt"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"go.uber.org/zap"
	"runtime"
	"time"
)

type RetriableFunc func(args ...any) error

func OnErr(
	ctx context.Context, tErrors []error, retries []time.Duration, fn RetriableFunc, fnArgs ...any) error {
	var retryCount int
	fnName := GetFunctionName()

	for {
		err := fn(fnArgs...)
		if err == nil {
			logger.Log.Debug("successfully executed", zap.String("function", fnName))
			return nil
		}

		var matchErr bool
		var tErr error
		for _, tErr = range tErrors {
			if errors.Is(err, tErr) {
				matchErr = true
				break
			}
		}
		if !matchErr {
			return err
		}

		if len(retries) == 0 {
			logger.Log.Debug("no more retries", zap.String("function", fnName))
			return err
		}

		retryCount++
		logger.Log.Debug("retrying",
			zap.String("fn", fnName),
			zap.String("error", tErr.Error()),
			zap.Int("retry_id", retryCount),
			zap.Duration("interval", retries[0]),
		)

		select {
		case <-time.After(retries[0]):
			retries = retries[1:]
		case <-ctx.Done():
			return fmt.Errorf("context cancelled for %s", fnName)
		}
	}
}

func GetFunctionName() string {
	fnName, _, _, _ := runtime.Caller(2)
	return fmt.Sprintf("%s()", runtime.FuncForPC(fnName).Name())
}
