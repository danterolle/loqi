package tui

import (
	"context"
	"fmt"

	"github.com/danterolle/voca/translate"
)

type CLIUI struct {
	Source string
	Target string
	Text   string
}

func NewCLIUI(source, target, text string) *CLIUI {
	return &CLIUI{Source: source, Target: target, Text: text}
}

func (u *CLIUI) Run(ctx context.Context, core *translate.Core) error {
	result, err := core.Backend.Translate(ctx, u.Text, u.Source, u.Target)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
