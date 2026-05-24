// Package components is the single source of truth for the
// `aikata add <component>` registry. Each Component knows how to add
// itself to an existing aikata project; the cli layer dispatches via
// the registry, and `aikata init --with-*` reuses the same renderers
// so flag-time and add-time behavior cannot drift.
//
// Layering rule: components depends on templates and config; nothing
// in scaffold/, cli/, or doctor/ may be imported from here. This keeps
// the dependency direction one-way:
//
//	cli → components → {templates, config, adr}
//	cli → scaffold   → {components, templates, config}
//
// See plan v0.4 for the component roster and rollout split.
package components
