package workcli

import (
	"context"
	"fmt"

	"github.com/gh-xj/agent-repo-kit/cli/internal/work"
)

type TriageCmd struct {
	Accept TriageAcceptCmd `cmd:"" help:"accept an inbox item"`
}

type TriageAcceptCmd struct {
	ID          string   `arg:"" help:"inbox item id"`
	Title       string   `help:"work item title override"`
	Description string   `help:"work item description override"`
	Status      string   `help:"initial work status" enum:"ready,active,blocked,done,cancelled" default:"ready"`
	Priority    string   `help:"priority label"`
	Area        string   `help:"work area"`
	Labels      []string `name:"label" help:"label to attach; repeatable"`
}

func (c *TriageAcceptCmd) Run(globals *CLI) error {
	store, err := globals.workStore()
	if err != nil {
		return err
	}
	item, err := store.AcceptInboxItem(context.Background(), acceptInboxItemInput{
		ID: c.ID,
		Options: work.AcceptInboxOptions{
			Title:       c.Title,
			Description: c.Description,
			Status:      work.WorkStatus(c.Status),
			Priority:    c.Priority,
			Area:        c.Area,
			Labels:      c.Labels,
		},
	})
	if err != nil {
		return err
	}
	out := globals.stdout()
	if globals.JSON {
		return emitJSON(out, map[string]any{
			"store": globals.Store,
			"item":  item,
		})
	}
	if id := fieldString(item, "ID", "Id"); id != "" {
		_, err = fmt.Fprintf(out, "accepted %s\n", id)
		return err
	}
	return printRecord(out, item)
}
