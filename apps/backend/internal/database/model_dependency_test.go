package database

import (
	"sort"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

// TestMigrationModelForeignKeyCycles is a database-free guard for the class
// of bug behind the "relation users does not exist" startup failure.
//
// GORM can only create tables one at a time, and it embeds each foreign key
// inline in CREATE TABLE. If two or more models reference each other
// (directly or through a chain), no creation order works on a real
// PostgreSQL database, but SQLite - which the rest of this package's tests
// use - does not enforce it, so such a cycle passes every SQLite test and
// then fails on the first deployment to an empty PostgreSQL database.
//
// The only cycle allowed is User <-> Organization, which
// MigrateUserOrganizationFoundation handles by staging its foreign keys.
// Introducing any other cycle fails this test until it is handled the same
// way (see user_organization_migration.go) and listed here deliberately.
func TestMigrationModelForeignKeyCycles(t *testing.T) {
	graph := map[string]map[string]bool{}
	cache := &sync.Map{}

	addEdge := func(from, to string) {
		if from == to {
			return // self-reference: a single table, creatable in one statement
		}
		if graph[from] == nil {
			graph[from] = map[string]bool{}
		}
		graph[from][to] = true
	}

	for _, model := range migrationModels() {
		parsed, err := schema.Parse(model, cache, schema.NamingStrategy{})
		if err != nil {
			t.Fatalf("parse model %T: %v", model, err)
		}
		if graph[parsed.Table] == nil {
			graph[parsed.Table] = map[string]bool{}
		}
		for _, rel := range parsed.Relationships.Relations {
			switch rel.Type {
			case schema.BelongsTo:
				// This table holds the foreign key, so it depends on the related table.
				addEdge(parsed.Table, rel.FieldSchema.Table)
			case schema.HasOne, schema.HasMany:
				// The related table holds the foreign key, so it depends on this one.
				addEdge(rel.FieldSchema.Table, parsed.Table)
			}
		}
	}

	var cycles []string
	for _, component := range stronglyConnectedComponents(graph) {
		if len(component) > 1 {
			sort.Strings(component)
			cycles = append(cycles, strings.Join(component, " <-> "))
		}
	}

	if len(cycles) != 1 || cycles[0] != "organizations <-> users" {
		t.Fatalf("expected exactly one foreign-key cycle (organizations <-> users, staged by MigrateUserOrganizationFoundation); found %v", cycles)
	}
}

// stronglyConnectedComponents is Tarjan's algorithm over a table-dependency graph.
func stronglyConnectedComponents(graph map[string]map[string]bool) [][]string {
	var (
		index    = 0
		stack    []string
		onStack  = map[string]bool{}
		indices  = map[string]int{}
		lowlinks = map[string]int{}
		result   [][]string
	)

	var visit func(node string)
	visit = func(node string) {
		indices[node] = index
		lowlinks[node] = index
		index++
		stack = append(stack, node)
		onStack[node] = true

		for next := range graph[node] {
			if _, seen := indices[next]; !seen {
				visit(next)
				if lowlinks[next] < lowlinks[node] {
					lowlinks[node] = lowlinks[next]
				}
			} else if onStack[next] && indices[next] < lowlinks[node] {
				lowlinks[node] = indices[next]
			}
		}

		if lowlinks[node] == indices[node] {
			var component []string
			for {
				last := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[last] = false
				component = append(component, last)
				if last == node {
					break
				}
			}
			result = append(result, component)
		}
	}

	nodes := make([]string, 0, len(graph))
	for node := range graph {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	for _, node := range nodes {
		if _, seen := indices[node]; !seen {
			visit(node)
		}
	}
	return result
}
