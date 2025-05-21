package app

import (
	"testing"

	"mark/internal/domain"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestContextItemsList(t *testing.T) {
	t.Parallel()

	t.Run("no items", func(t *testing.T) {
		t.Parallel()

		cil := NewContextItemsList()
		cil.SetSize(20, 10)
		v := cil.View()
		snaps.MatchStandaloneSnapshot(t, v)
	})

	t.Run("with items", func(t *testing.T) {
		t.Parallel()

		t.Run("without focus", func(t *testing.T) {
			t.Parallel()

			cil := NewContextItemsList()
			cil.SetSize(20, 10)
			cil.SetItemsFromSessionContextItems([]domain.ContextItem{domain.TextItem("item 1"), domain.TextItem("item 2"), domain.TextItem("item 3")})
			cil.Blur()
			v := cil.View()
			snaps.MatchStandaloneSnapshot(t, v)
		})

		t.Run("with focus", func(t *testing.T) {
			t.Parallel()

			cil := NewContextItemsList()
			cil.SetSize(20, 10)
			cil.SetItemsFromSessionContextItems([]domain.ContextItem{domain.TextItem("item 1"), domain.TextItem("item 2"), domain.TextItem("item 3")})
			cil.Focus()
			v := cil.View()
			snaps.MatchStandaloneSnapshot(t, v)
		})
	})
}
