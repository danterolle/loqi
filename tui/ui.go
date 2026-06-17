package tui

import (
	"context"

	"github.com/danterolle/voca/translate"
)

type UI interface {
	Run(ctx context.Context, core *translate.Core) error
}
