package tui

import (
	"context"
	"fmt"

	"github.com/danterolle/voca/translate"
)

func RunCLI(ctx context.Context, core *translate.Core, source, target, text string) error {
	result, err := core.Translate(ctx, text, source, target)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
