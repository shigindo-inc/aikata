package components

import (
	"io"
	"time"

	"github.com/shigindo-inc/aikata/internal/templates"
)

// Component is the contract every entry in the registry implements.
// The interface is deliberately narrow: a name + metadata for
// discoverability, and Add for the authoring operation. Argument
// validation lives inside Add so each component can shape its own
// errors without forcing a one-size-fits-all spec.
type Component interface {
	Name() string
	Description() string
	Status() string
	// Add executes the component's authoring operation. Contract:
	//   - failure must leave the filesystem unchanged (all-or-nothing)
	//   - idempotent when feasible (return nil + notice on Stderr when
	//     there is nothing to do)
	//   - dry-run support is the caller's responsibility via ctx.DryRun
	Add(ctx AddContext) error
}

// Status values used by Component.Status. StatusActive marks a
// component the registry can apply today; StatusReserved holds a name
// for a future release.
const (
	StatusActive   = "active"
	StatusReserved = "reserved"
)

// AddContext carries the inputs every Component.Add invocation needs.
// Required fields:
//
//   - TargetDir — absolute project root
//   - ProjectName — value of project.name in .aikata/aikata.yaml
//
// Optional fields:
//
//   - Lang — "en" / "ja"; empty defaults to "en"
//   - Clock — time source for template helpers; nil = time.Now
//   - Args — positional arguments after `aikata add <name>`
//   - DryRun — print plan to Stdout without writing
//   - Stdout / Stderr — message sinks; nil falls back to os.Stdout / os.Stderr
type AddContext struct {
	TargetDir   string
	ProjectName string
	Lang        string
	Clock       templates.Clock
	Args        []string
	DryRun      bool
	Stdout      io.Writer
	Stderr      io.Writer
}

// Now returns the timestamp for this Add invocation, derived from
// AddContext.Clock or time.Now when Clock is nil.
func (c AddContext) Now() time.Time {
	if c.Clock != nil {
		return c.Clock()
	}
	return time.Now()
}
