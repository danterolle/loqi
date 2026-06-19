package ollama

import (
	"github.com/danterolle/voca/translate"
)

var _ translate.Backend = (*Backend)(nil)
