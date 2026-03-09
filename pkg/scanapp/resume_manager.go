package scanapp

import (
	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/state"
)

func persistResumeState(cfg config.Config, opts RunOptions, logger *scanLogger, runtimes []*chunkRuntime, dispatchErr, runErr error) error {
	incomplete := hasIncomplete(runtimes)
	if !incomplete && runErr == nil && !shouldSaveOnDispatchErr(dispatchErr) {
		return nil
	}

	savePath := resumePath(cfg, opts)
	if err := state.Save(savePath, collectChunkStates(runtimes)); err != nil {
		return err
	}
	logger.infof("resume state saved to %s", savePath)
	return nil
}
