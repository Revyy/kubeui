package podview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func getModel() Model {
	return New(v1.Pod{}, 100)
}

func applyKeyMsgToNewModel(model Model, keyType tea.KeyType) (Model, tea.Cmd) {
	return model.Update(tea.KeyMsg{Type: keyType})
}

func TestModel_Update(t *testing.T) {
	// KEYS
	t.Run("Test keystrokes", func(t *testing.T) {
		t.Parallel()
		t.Run("Move cursor right", func(t *testing.T) {
			t.Parallel()
			model, cmd := applyKeyMsgToNewModel(getModel(), tea.KeyRight)
			assert.Nil(t, cmd, "expected command to be nil")
			assert.Equal(t, 1, model.cursor, "expected cursor to increment")
			assert.Equal(t, ANNOTATIONS, model.view, "view should be changed to ANNOTATIONS")

			// Cursor moving past the end.
			model.cursor = len(model.views) - 1
			model, cmd = applyKeyMsgToNewModel(model, tea.KeyRight)
			assert.Nil(t, cmd, "expected command to be nil")
			assert.Equal(t, 0, model.cursor, "expected cursor to reset when reaching the end")
			assert.Equal(t, STATUS, model.view, "view should be changed to STATUS")

		})

		t.Run("Move cursor left", func(t *testing.T) {
			t.Parallel()
			model := getModel()
			model.cursor = 1
			model.view = ANNOTATIONS

			model, cmd := applyKeyMsgToNewModel(model, tea.KeyLeft)
			assert.Nil(t, cmd, "expected command to be nil")
			assert.Equal(t, 0, model.cursor, "expected cursor to decrement")
			assert.Equal(t, STATUS, model.view, "view should be changed to STATUS")

			// Cursor should wrap around to the last position if we go below 0.
			model, cmd = applyKeyMsgToNewModel(getModel(), tea.KeyLeft)
			assert.Nil(t, cmd, "expected command to be nil")
			assert.Equal(t, len(model.views)-1, model.cursor, "expected cursor to wrap to end")
			assert.Equal(t, LABELS, model.view, "view should be changed to LABELS")
		})

	})
}
