package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	overlay "github.com/floatpane/bubble-overlay"
)

type page int

const (
	pageMenu page = iota
	pageAbout
	pageSkills
	pageContact
)

var menuItems = []string{"About", "Skills", "Contact", "Quitter"}

const asciiPlaceholder = `..........=*********=:.........
......-+****#%%%%%#*****-......
....=*****#%@@@%@@@%%#****=....
..-******#@@@@%%%#*#%@%*****...
.=#******%@%@%%*-==-=%@#*****..
:##*****#%%%%##=-==-=+%#******.
*##*****#@%##*+++===*#%#******=
###*****#%%#+=++=-+**#%#*******
###******##*=---==+==##********
###*******#*==--=-==+*#********
+##*********+==++**+*#****###%=
.###********===-==****#**##%@%.
..##******#+==-==+*@@@@@@%@@%..
...****%@@@%+====+#@@@@@%@@#...
....=%@@@@@@@%#*%@@@@@@@%%-....
......:%@@@@@@@@@@@@@@@%.......
..........*@@@@@@@@@+..........`

var (
	accent = lipgloss.Color("#5eead4") // cyan glow
	muted  = lipgloss.Color("#4f8a83")
	text   = lipgloss.Color("#bdf5ea")
	dimBg  = lipgloss.Color("#123334") // pour les bordures discrètes

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	selectedStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(muted)
	helpStyle     = lipgloss.NewStyle().Foreground(muted)

	chromeStyle = lipgloss.NewStyle().Foreground(muted)
	headerStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)

	sideBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Foreground(text).
			Padding(1, 2).
			Width(37)

	contentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimBg).
			Foreground(text).
			Padding(1, 3).
			Width(70).
			Height(25).
			MaxHeight(25) // jamais plus haut que le cadre, sinon la frame dépasse le terminal

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Foreground(text).
			Padding(1, 4).
			Align(lipgloss.Center)

	// Une seule ligne par badge (pas de bordure) pour garder la page Skills courte.
	badgeStyle = lipgloss.NewStyle().
			Background(dimBg).
			Foreground(accent).
			Padding(0, 1)
)

// contentWidth est la largeur réellement disponible pour le texte dans
// contentStyle (Width inclut bordure et padding en Lip Gloss v2).
func contentWidth() int {
	return contentStyle.GetWidth() - contentStyle.GetHorizontalFrameSize()
}

type model struct {
	width       int
	height      int
	cursor      int
	page        page
	confirmQuit bool
}

func wrapBadges(items ...string) string {
	const sep = "  "
	maxWidth := contentWidth()

	var rows []string
	var current string
	currentWidth := 0

	for _, item := range items {
		badge := badgeStyle.Render(item)
		w := lipgloss.Width(badge)

		if currentWidth > 0 && currentWidth+len(sep)+w > maxWidth {
			rows = append(rows, current)
			current, currentWidth = "", 0
		}

		if currentWidth > 0 {
			current += sep
			currentWidth += len(sep)
		}
		current += badge
		currentWidth += w
	}

	if currentWidth > 0 {
		rows = append(rows, current)
	}

	return strings.Join(rows, "\n\n") // ligne vide entre les rangées
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		if m.confirmQuit {
			switch msg.String() {
			case "y", "Y", "ctrl+c", "enter":
				return m, tea.Quit

			case "n", "N", "esc":
				m.confirmQuit = false

			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q", "esc":
			if m.page == pageMenu {
				return m, tea.Quit
			}

			m.page = pageMenu

		case "up", "k":
			if m.page == pageMenu && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.page == pageMenu && m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter":
			if m.page == pageMenu {
				switch m.cursor {
				case 0:
					m.page = pageAbout
				case 1:
					m.page = pageSkills
				case 2:
					m.page = pageContact
				case 3:
					m.confirmQuit = true
				}
			}
		}
	}
	return m, nil

}

func (m model) View() tea.View {
	var body string

	switch m.page {
	case pageAbout:
		body = titleStyle.Render("À propos") + "\n\n" +
			"Étudiante en Master 1 Industrie Numérique à Rennes 1,\n" +
			"en alternance chez AMJ GROUPE. Développement fullstack :\n" +
			"Angular côté frontend, Spring Boot côté backend.\n\n" +
			"En ce moment : projets perso pour explorer de nouvelles\n" +
			"technos, et apprentissage des pipelines CI/CD.\n\n" +
			helpStyle.Render("[esc/q] retour")

	case pageSkills:
		body = titleStyle.Render("Skills") + "\n\n" +
			normalStyle.Render("BACKEND") + "\n" +
			wrapBadges("Java", "Spring Boot", "Node.js", "Express.js", "Hibernate", "JWT") + "\n\n" +
			normalStyle.Render("FRONTEND") + "\n" +
			wrapBadges("Angular", "React", "TypeScript", "JavaScript", "HTML5", "CSS3", "TailwindCSS", "Bootstrap") + "\n\n" +
			normalStyle.Render("DATA") + "\n" +
			wrapBadges("MySQL", "PostgreSQL", "MongoDB", "Firebase") + "\n\n" +
			normalStyle.Render("OUTILS & AUTRES") + "\n" +
			wrapBadges("Git", "Figma", "Jest", "Go", "Bubble Tea") + "\n\n" +
			helpStyle.Render("[esc/q] retour")

	case pageContact:
		body = titleStyle.Render("Contact") + "\n\n" +
			"Email : syncik.dev@gmail.com\n" +
			"GitHub : github.com/ikouRe\n\n" +
			helpStyle.Render("[esc/q] retour")

	default: // pageMenu
		body = titleStyle.Render("TUIfolio") + "\n\n"
		for i, item := range menuItems {
			cursor := "  "
			style := normalStyle
			if i == m.cursor {
				cursor = "> "
				style = selectedStyle
			}
			body += cursor + style.Render(item) + "\n"
		}
		body += "\n" + helpStyle.Render("[up/down ou j/k] naviguer  [enter] sélectionner  [q] quitter")
	}

	sidebar := sideBarStyle.Render(asciiPlaceholder + "\n\nSyncik\nDev Fullstack\nSpringBoot · Angular ")
	content := contentStyle.Render(body)
	middle := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	chrome := chromeStyle.Render("● ● ●  syncik :: profile_scan.sh")
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		headerStyle.Render("PROFILE SCAN"),
		lipgloss.NewStyle().Width(40).Render(""), // espaceur
		normalStyle.Render("● STATUS: ONLINE"),
	)

	layout := lipgloss.JoinVertical(lipgloss.Left, chrome, "", header, "", middle)

	if m.height > 0 && lipgloss.Height(layout) > m.height {
		layout = lipgloss.JoinVertical(lipgloss.Left, chrome, header, middle)
	}
	if m.height > 0 && lipgloss.Height(layout) > m.height {
		layout = strings.Join(strings.Split(layout, "\n")[:m.height], "\n")
	}

	if m.confirmQuit {
		popup := popupStyle.Render(
			titleStyle.Render("Êtes-vous sûr de vouloir quitter ? (y/n)") + "\n\n" + helpStyle.Render("[y] oui  [n] non"))
		layout = overlay.Center(layout, popup, m.width, m.height)
	}
	v := tea.NewView(layout)
	v.AltScreen = true
	return v
}
