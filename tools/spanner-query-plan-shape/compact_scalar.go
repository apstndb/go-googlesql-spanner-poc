package main

import (
	"fmt"
	"strings"

	"cloud.google.com/go/spanner/apiv1/spannerpb"
)

type compactHiddenScalarChildGroup struct {
	displayName string
	linkTypes   []string
}

func compactHiddenScalarChildAnnotations(node *spannerpb.PlanNode, nodesByIndex map[int32]*spannerpb.PlanNode) []string {
	groups := compactHiddenScalarChildGroups(node, nodesByIndex)
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		out = append(out, group.annotation())
	}
	return out
}

func compactHiddenScalarChildGroups(node *spannerpb.PlanNode, nodesByIndex map[int32]*spannerpb.PlanNode) []compactHiddenScalarChildGroup {
	if node == nil {
		return nil
	}

	visibleChildIndexes := map[int32]bool{}
	for _, link := range node.GetChildLinks() {
		if compactTreeLinkVisible(link, nodesByIndex) {
			visibleChildIndexes[link.GetChildIndex()] = true
		}
	}

	var groups []compactHiddenScalarChildGroup
	groupByDisplayName := map[string]int{}
	seen := map[string]bool{}
	for _, link := range node.GetChildLinks() {
		child := nodesByIndex[link.GetChildIndex()]
		if child == nil {
			continue
		}
		if visibleChildIndexes[link.GetChildIndex()] {
			continue
		}

		linkType := strings.TrimSpace(link.GetType())
		displayName := compactHiddenScalarChildDisplayName(child)
		seenKey := displayName + "\x00" + linkType
		if seen[seenKey] {
			continue
		}
		seen[seenKey] = true

		groupIndex, ok := groupByDisplayName[displayName]
		if !ok {
			groupIndex = len(groups)
			groupByDisplayName[displayName] = groupIndex
			groups = append(groups, compactHiddenScalarChildGroup{
				displayName: displayName,
				linkTypes:   []string{},
			})
		}
		groups[groupIndex].linkTypes = append(groups[groupIndex].linkTypes, linkType)
	}
	return groups
}

func compactHiddenScalarChildDisplayName(child *spannerpb.PlanNode) string {
	displayName := strings.TrimSpace(child.GetDisplayName())
	if displayName == "" {
		return "Unknown"
	}
	return displayName
}

func (g compactHiddenScalarChildGroup) annotation() string {
	if len(g.linkTypes) == 1 && (g.linkTypes[0] == "" || g.linkTypes[0] == g.displayName) {
		return g.displayName
	}
	return fmt.Sprintf("%s[%s]", g.displayName, strings.Join(g.linkTypes, ", "))
}
