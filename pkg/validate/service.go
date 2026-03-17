package validate

import (
	"fmt"
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
)

type Result struct {
	Valid  bool
	Detail string
}

func Inputs(cfg config.Config) Result {
	cidrFile, err := os.Open(cfg.CIDRFile)
	if err != nil {
		return Result{Valid: false, Detail: fmt.Sprintf("failed to open cidr file: %v", err)}
	}
	defer cidrFile.Close()

	if _, err := input.LoadCIDRsWithColumns(cidrFile, cfg.CIDRIPCol, cfg.CIDRIPCidrCol); err != nil {
		return Result{Valid: false, Detail: err.Error()}
	}

	portFile, err := os.Open(cfg.PortFile)
	if err != nil {
		return Result{Valid: false, Detail: fmt.Sprintf("failed to open port file: %v", err)}
	}
	defer portFile.Close()

	if _, err := input.LoadPorts(portFile); err != nil {
		return Result{Valid: false, Detail: err.Error()}
	}

	return Result{Valid: true, Detail: "ok"}
}
