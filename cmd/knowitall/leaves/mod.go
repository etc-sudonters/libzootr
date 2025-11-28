package leaves

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

const VERSION = 1

func UnwrapStatusMessage(cmd tea.Cmd) (StatusMsg, error) {
	model := cmd()
	msg, ok := model.(StatusMsg)
	if ok {
		return msg, nil
	}
	return "", errors.New("did not produce status msg")
}
