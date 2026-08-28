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

	tagStyle := lipgloss.NewStyle().
		Foreground(m.theme.Highlight).
		Background(m.theme.TabActiveBg).
		Padding(0, 1)

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

			statusIcon := "[ ]"
			if t.IsDone {
				statusIcon = "[x]"
			} else if t.Title == m.snapshot.ActiveTaskTitle {
				statusIcon = "[*]"
			}

			// Estimation dots: e.g. [●●○] (2/3)
			var estIcons strings.Builder
			for p := 1; p <= t.Target; p++ {
				if p <= t.Completed {
					estIcons.WriteString("●")
				} else {
					estIcons.WriteString("○")
				}
			}
			estStr := fmt.Sprintf("[%s] (%d/%d)", estIcons.String(), t.Completed, t.Target)

			// Render Tags
			var renderedTags string
			if len(t.Tags) > 0 {
				var tagPills []string
				for _, tag := range t.Tags {
					tagPills = append(tagPills, tagStyle.Render("#"+tag))
				}
				renderedTags = " " + strings.Join(tagPills, " ")
			}

			title := t.Title
			maxTitleLen := contentWidth - 42
			if maxTitleLen > 10 && len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}

			line := fmt.Sprintf("%s%s %-26s %s %s", cursor, statusIcon, title, renderedTags, estStr)
			if t.Title == m.snapshot.ActiveTaskTitle && !t.IsDone {
				line += " " + lipgloss.NewStyle().Bold(true).Foreground(m.theme.Accent).Render("*ACTIVE*")
			}

			if t.IsDone {
				content.WriteString(taskDoneStyle.Render(line) + "\n")
			} else if i == m.selectedTask {
				content.WriteString(taskActiveStyle.Render(line) + "\n")
			} else {
				content.WriteString(lipgloss.NewStyle().Foreground(m.theme.Text).Render(line) + "\n")
			}
		}
	}

	if m.inputMode == ModeAddingTask {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).Render("Add Task: ") + m.textInput.View())
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.TextDim).Render("Tip: include tags and estimates, e.g. Refactor API #backend est:3  [Enter/Esc]"))
	} else if m.inputMode == ModeEditingTask {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.Highlight).Bold(true).Render("Edit Task: ") + m.textInput.View())
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.TextDim).Render("Edit title, tags (#tag), or estimates (est:N)  [Enter/Esc]"))
	} else {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(m.theme.BorderActive).Render(strings.Repeat("-", contentWidth-6)) + "\n")
		content.WriteString(lipgloss.NewStyle().Foreground(m.theme.TextDim).Render("[j/k] Navigate   [J/K] Reorder   [Space] Set Active   [a] Add   [e] Edit   [d] Done   [x] Del   [C] Clear Done"))
	}

	return m.makeBox(contentWidth, calcHeight, lipgloss.Left, content.String())
}
