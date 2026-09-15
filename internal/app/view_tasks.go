package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderTasksTab renders Tab 2: Vim-style Task Queue, Tags, and Estimation Manager.
func (m Model) renderTasksTab(contentWidth, termH int) string {
	calcHeight := termH - 7
	if calcHeight < 8 {
		calcHeight = 8
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Text)

	taskActiveStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Accent)

	taskDoneStyle := lipgloss.NewStyle().
		Foreground(m.theme.TextDim)

	// Plain foreground (no padding/background chip): the tags column has a fixed width, and
	// per-tag padding would push the total row width past that budget and break alignment.
	tagStyle := lipgloss.NewStyle().
		Foreground(m.theme.Highlight)

	var content strings.Builder
	content.WriteString(headerStyle.Render("TASKS (Vim Mode)") + "\n\n")

	if len(m.tasks) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(m.theme.TextDim).Render("  No tasks in queue. Press 'a' to add a task.\n\n"))
	} else {
		maxItems := calcHeight - 5
		if maxItems < 2 {
			maxItems = 2
		}

		startIdx := 0
		if m.selectedTask >= maxItems {
			startIdx = m.selectedTask - maxItems + 1
		}
		endIdx := startIdx + maxItems
		if endIdx > len(m.tasks) {
			endIdx = len(m.tasks)
		}

		for i := startIdx; i < endIdx; i++ {
			t := m.tasks[i]
			cursor := "  "
			if i == m.selectedTask {
				cursor = "> "
			}

			isActiveTask := !t.IsDone && t.Title == m.snapshot.ActiveTaskTitle

			statusIcon := "[ ]"
			if t.IsDone {
				statusIcon = "[x]"
			} else if isActiveTask {
				statusIcon = "[*]"
			}

			// Fixed-width columns so every row occupies exactly one line and lines up like a
			// table, regardless of title/tag length or how large a task's pomo target is.
			const progressColWidth = 15
			const tagsColWidth = 18
			usable := boxContentWidth(contentWidth)
			// cursor(2) + icon(3) + spacing(2, before progress and before title) + progress column
			baseOverhead := 2 + 3 + 2 + progressColWidth
			showTagsCol := usable-baseOverhead-1-tagsColWidth >= 10
			titleColWidth := usable - baseOverhead
			if showTagsCol {
				titleColWidth -= 1 + tagsColWidth // extra separator + the tags column itself
			}
			if titleColWidth < 6 {
				titleColWidth = 6
			}

			// Progress column: dot icons only for small targets, otherwise a plain "n/m"
			// fraction — a target of 50 would otherwise render 50 dots and blow out the row.
			var progressText string
			if t.Target <= 8 {
				var dots strings.Builder
				for p := 1; p <= t.Target; p++ {
					if p <= t.Completed {
						dots.WriteString("●")
					} else {
						dots.WriteString("○")
					}
				}
				progressText = fmt.Sprintf("[%s](%d/%d)", dots.String(), t.Completed, t.Target)
			} else {
				progressText = fmt.Sprintf("(%d/%d)", t.Completed, t.Target)
			}

			tagsText := ""
			if len(t.Tags) > 0 {
				var withHash []string
				for _, tag := range t.Tags {
					withHash = append(withHash, "#"+tag)
				}
				tagsText = strings.Join(withHash, " ")
			}

			line := cursor + statusIcon + " " + fitRunes(t.Title, titleColWidth)
			if showTagsCol {
				line += " " + tagStyle.Render(fitRunes(tagsText, tagsColWidth))
			}
			line += " " + fitRunes(progressText, progressColWidth)

			// Guarantee the row is exactly `usable` columns even if a styled tag pill's
			// background color padding nudged the width — never let a row wrap.
			if w := lipgloss.Width(line); w > usable {
				line = lipgloss.NewStyle().MaxWidth(usable).Render(line)
			}

			switch {
			case t.IsDone:
				content.WriteString(taskDoneStyle.Render(line) + "\n")
			case i == m.selectedTask || isActiveTask:
				content.WriteString(taskActiveStyle.Render(line) + "\n")
			default:
				content.WriteString(lipgloss.NewStyle().Foreground(m.theme.Text).Render(line) + "\n")
			}
		}
	}

	usableBottom := boxContentWidth(contentWidth)
	if m.inputMode == ModeAddingTask {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).Render("Add Task: ") + m.textInput.View())
		tip := truncateRunes("Tip: include tags and estimates, e.g. Refactor API #backend est:3  [Enter/Esc]", usableBottom)
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.TextDim).Render(tip))
	} else if m.inputMode == ModeEditingTask {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.Highlight).Bold(true).Render("Edit Task: ") + m.textInput.View())
		tip := truncateRunes("Edit title, tags (#tag), or estimates (est:N)  [Enter/Esc]", usableBottom)
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.TextDim).Render(tip))
	} else {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.BorderActive).Render(strings.Repeat("-", contentWidth-6)) + "\n")
		hints := []string{"[j/k] Navigate", "[J/K] Reorder", "[Space] Set Active", "[a] Add", "[e] Edit", "[d] Done", "[x] Del", "[C] Clear Done"}
		for i, h := range hints {
			hints[i] = lipgloss.NewStyle().Foreground(m.theme.TextDim).Render(h)
		}
		content.WriteString(wrapItems(hints, usableBottom, "   "))
	}

	return m.makeBox(contentWidth, calcHeight, lipgloss.Left, content.String())
}
