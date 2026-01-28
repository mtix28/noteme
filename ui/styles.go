package ui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#7D56F4") // Purple
	secondaryColor = lipgloss.Color("#FF79C6") // Pink
	subtleColor    = lipgloss.Color("#6272A4") // Gray-ish
	accentColor    = lipgloss.Color("#50FA7B") // Green
    textColor      = lipgloss.Color("#F8F8F2") // White
    
    // Heatmap Colors (activity levels)
    heatLevel0 = lipgloss.Color("#282A36") // None
    heatLevel1 = lipgloss.Color("#44475A") // Low
    heatLevel2 = lipgloss.Color("#6272A4") // Medium
    heatLevel3 = lipgloss.Color("#BD93F9") // High
    heatLevel4 = lipgloss.Color("#FF79C6") // Max

	// App container
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	// Headers
	titleStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
        MarginBottom(1)
    
    headerStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(primaryColor).
        Padding(1, 2).
        MarginBottom(1)

	// Dashboard Cards
	cardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(subtleColor).
		Padding(1, 2).
        MarginRight(1)

	// Text Styles
	helpStyle = lipgloss.NewStyle().
		Foreground(subtleColor).
		MarginTop(1)
    
    statLabel = lipgloss.NewStyle().
        Foreground(subtleColor)
        
    statValue = lipgloss.NewStyle().
        Foreground(accentColor).
        Bold(true).
        MarginLeft(1)
        
    emptyStateStyle = lipgloss.NewStyle().
        Foreground(subtleColor).
        Italic(true).
        Align(lipgloss.Center)
)