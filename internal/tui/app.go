package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"macsetup/internal/config"
	"macsetup/internal/installer"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppState int

const (
	StateWelcome AppState = iota
	StateScanning
	StateSelection
	StateXcodeWait
	StateInstalling
	StateSummary
)

type Model struct {
	ctx       context.Context
	state     AppState
	width     int
	height    int
	workers   int
	verbose   bool
	err       error
	startTime time.Time

	categories []config.Category
	packages   []config.Package

	installed map[string]bool
	selected  map[string]bool
	collapsed map[string]bool

	cursor       int
	scrollOffset int
	listItems    []listItem

	spin spinner.Model
	bar  progress.Model

	progressUpdates <-chan installer.ProgressUpdate
	installDoneCh   <-chan installer.Summary
	installErrCh    <-chan error
	results         []installer.InstallResult

	totalPackages     int
	completedPackages int

	// Track packages by state
	installedPackages map[string]string // name -> message
	failedPackages    map[string]string // name -> error
	runningPackages   map[string]string // name -> message
}

type listItem struct {
	isCategory      bool
	isSubCategory   bool
	subCategoryName string
	category        *config.Category
	pkg             *config.Package
}

type (
	installDoneMsg    installer.Summary
	xcodeReadyMsg     struct{}
	scanFinishedMsg   map[string]bool
	installStartedMsg struct {
		updates <-chan installer.ProgressUpdate
		done    <-chan installer.Summary
		errs    <-chan error
	}
)
type errMsg struct{ err error }

func Run(ctx context.Context, workers int, verbose bool) error {
	m := newModel(ctx, workers, verbose)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(ctx))
	_, err := p.Run()
	return err
}

func newModel(ctx context.Context, workers int, verbose bool) Model {
	spin := spinner.New()
	spin.Spinner = spinner.Dot

	bar := progress.New(progress.WithSolidFill("#8cc265"))

	m := Model{
		ctx:               ctx,
		state:             StateWelcome,
		workers:           workers,
		verbose:           verbose,
		categories:        config.Categories(),
		packages:          config.AllPackages(),
		selected:          config.DefaultSelection(),
		collapsed:         make(map[string]bool),
		spin:              spin,
		bar:               bar,
		installedPackages: make(map[string]string),
		failedPackages:    make(map[string]string),
		runningPackages:   make(map[string]string),
	}

	// Collapse all categories by default
	for _, cat := range m.categories {
		m.collapsed[cat.Key] = true
	}

	m.rebuildList()
	return m
}

func (m Model) Init() tea.Cmd {
	return m.spin.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case installer.ProgressUpdate:
		m = m.applyUpdate(msg)
		return m, m.waitForUpdate()
	case installStartedMsg:
		m.progressUpdates = msg.updates
		m.installDoneCh = msg.done
		m.installErrCh = msg.errs
		// Count total packages to install
		m.totalPackages = 0
		for _, selected := range m.selected {
			if selected {
				m.totalPackages++
			}
		}
		m.completedPackages = 0
		return m, tea.Batch(m.waitForUpdate(), m.waitForDone(), m.spin.Tick)
	case installDoneMsg:
		m.state = StateSummary
		m.results = msg.Results
		return m, nil
	case errMsg:
		m.err = msg.err
		m.state = StateSummary
		return m, nil
	case xcodeReadyMsg:
		m.state = StateInstalling
		return m, m.startInstall()
	case scanFinishedMsg:
		m.installed = msg
		// Unselect installed packages if not required
		for pkgName := range m.installed {
			// Find pkg
			var pkg *config.Package
			for i := range m.packages {
				if m.packages[i].Name == pkgName {
					pkg = &m.packages[i]
					break
				}
			}
			if pkg != nil && !pkg.Required && pkg.Category != "core" {
				m.selected[pkgName] = false
			}
		}
		m.state = StateSelection
		return m, nil
	}

	if m.state == StateInstalling || m.state == StateScanning {
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	switch m.state {
	case StateWelcome:
		if msg.String() == "enter" {
			m.state = StateScanning
			return m, m.startScan()
		}
	case StateSelection:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateScroll()
			}
		case "down", "j":
			if m.cursor < len(m.listItems)-1 {
				m.cursor++
				m.updateScroll()
			}
		case " ":
			m.toggleCurrent()
		case "a":
			m.selectCategoryAtCursor(true)
		case "n":
			m.selectCategoryAtCursor(false)
		case "enter":
			m.startTime = time.Now()
			if installer.IsXcodeInstalled(m.ctx) {
				m.state = StateInstalling
				return m, m.startInstall()
			}
			m.state = StateXcodeWait
			_ = installer.TriggerXcodeInstall(m.ctx)
			return m, m.waitForXcode()
		}
	case StateXcodeWait:
		switch msg.String() {
		case "enter":
			return m, nil
		}
	case StateSummary:
		switch msg.String() {
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) updateScroll() {
	viewHeight := m.height - 5
	if viewHeight < 1 {
		viewHeight = 5
	}

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+viewHeight {
		m.scrollOffset = m.cursor - viewHeight + 1
	}
}

func (m *Model) toggleCurrent() {
	if m.cursor < 0 || m.cursor >= len(m.listItems) {
		return
	}
	item := m.listItems[m.cursor]

	if item.isCategory && item.category != nil {
		catKey := item.category.Key
		m.collapsed[catKey] = !m.collapsed[catKey]
		m.rebuildList()
		return
	}

	if item.pkg == nil {
		return
	}
	if item.pkg.Required || item.pkg.Category == "core" || item.pkg.Category == "fonts" {
		return
	}
	key := item.pkg.Name
	m.selected[key] = !m.selected[key]
}

func (m *Model) selectCategoryAtCursor(on bool) {
	if m.cursor < 0 || m.cursor >= len(m.listItems) {
		return
	}
	item := m.listItems[m.cursor]
	var catKey string
	if item.isCategory && item.category != nil {
		catKey = item.category.Key
	} else if item.pkg != nil {
		catKey = item.pkg.Category
	} else {
		return
	}

	for _, c := range m.categories {
		if c.Key == catKey && (c.Required || !c.Selectable) {
			return
		}
	}

	for _, pkg := range m.packages {
		if pkg.Category != catKey {
			continue
		}
		if pkg.Required || pkg.Category == "core" || pkg.Category == "fonts" {
			continue
		}
		// If on=true (selecting all), we might select installed ones.
		// If the user explicitly hits 'a', they probably want to select everything in that category
		// regardless of installed status.
		// Or should we only select uninstalled ones?
		// Standard behavior: 'a' selects all.
		m.selected[pkg.Name] = on
	}
}

func (m *Model) rebuildList() {
	var items []listItem
	pkgsByCat := map[string][]config.Package{}

	// Separate packages into categories dynamically
	for _, p := range m.packages {
		// If installed and NOT selected (meaning we are skipping it), it goes to "Already Installed"
		// If installed and SELECTED (user wants to reinstall), it stays in original category
		// Exception: Required/Core/Fonts usually stay in place or we treat them normally?
		// User said "already installed packages should not be a part of individual categories below".
		// Let's adhere to that for non-required ones.
		// For required ones, they are always selected, so they stay in original category.

		effectiveCat := p.Category
		isInstalled := m.installed[p.Name]
		isSelected := m.selected[p.Name]

		if isInstalled && !isSelected {
			effectiveCat = "installed"
		}

		pkgsByCat[effectiveCat] = append(pkgsByCat[effectiveCat], p)
	}

	for _, cat := range m.categories {
		// Skip empty categories to reduce clutter? Or show empty ones?
		// Let's show them if they exist in config.

		// Skip core from list as per previous logic (it's hidden/auto-handled)
		if cat.Key == "core" {
			continue
		}

		pkgs := pkgsByCat[cat.Key]
		// If no packages in "installed" category, maybe hide it?
		if cat.Key == "installed" && len(pkgs) == 0 {
			continue
		}

		c := cat
		items = append(items, listItem{isCategory: true, category: &c})

		if m.collapsed[cat.Key] {
			continue
		}

		// Group by SubCategory
		subCats := map[string][]config.Package{}
		var subCatKeys []string
		for _, p := range pkgs {
			if _, exists := subCats[p.SubCategory]; !exists {
				subCatKeys = append(subCatKeys, p.SubCategory)
			}
			subCats[p.SubCategory] = append(subCats[p.SubCategory], p)
		}
		sort.Strings(subCatKeys)

		for _, sc := range subCatKeys {
			if sc != "" {
				items = append(items, listItem{isSubCategory: true, subCategoryName: sc})
			}
			pList := subCats[sc]
			sort.Slice(pList, func(i, j int) bool { return strings.Compare(pList[i].Name, pList[j].Name) < 0 })

			for i := range pList {
				p := pList[i]
				items = append(items, listItem{isCategory: false, pkg: &p})
			}
		}
	}
	m.listItems = items
	if m.cursor >= len(m.listItems) {
		m.cursor = len(m.listItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) startScan() tea.Cmd {
	return func() tea.Msg {
		installed, err := installer.ScanInstalledPackages(m.ctx, m.packages)
		if err != nil {
			// If scanning fails (e.g. brew not installed), we assume nothing installed or handle gracefully
			// For now, return empty map or partial results
			installed = make(map[string]bool)
		}
		if installed == nil {
			installed = make(map[string]bool)
		}
		return scanFinishedMsg(installed)
	}
}

func (m Model) startInstall() tea.Cmd {
	return func() tea.Msg {
		manager := installer.NewManager(m.workers, installer.RunOptions{Verbose: m.verbose})
		updates := manager.Progress()
		done := make(chan installer.Summary, 1)
		errs := make(chan error, 1)
		go func() {
			summary, err := manager.Run(m.ctx, m.selected)
			if err != nil {
				errs <- err
				return
			}
			done <- summary
		}()
		return installStartedMsg{updates: updates, done: done, errs: errs}
	}
}

func (m Model) waitForXcode() tea.Cmd {
	return func() tea.Msg {
		_ = installer.WaitForXcode(m.ctx, 2*time.Second)
		return xcodeReadyMsg{}
	}
}

func (m Model) applyUpdate(upd installer.ProgressUpdate) Model {
	pkgName := upd.Package.Name

	switch upd.Status {
	case installer.StatusRunning:
		// Remove from other states if present
		delete(m.installedPackages, pkgName)
		delete(m.failedPackages, pkgName)
		// Add to running
		m.runningPackages[pkgName] = upd.Message
	case installer.StatusInstalled, installer.StatusSkipped:
		// Remove from running
		delete(m.runningPackages, pkgName)
		// Add to installed
		m.installedPackages[pkgName] = upd.Message
		m.completedPackages++
	case installer.StatusFailed:
		// Remove from running
		delete(m.runningPackages, pkgName)
		// Add to failed
		m.failedPackages[pkgName] = upd.Error
		m.completedPackages++
	}
	return m
}

func (m Model) waitForUpdate() tea.Cmd {
	return func() tea.Msg {
		if m.progressUpdates == nil {
			return nil
		}
		upd, ok := <-m.progressUpdates
		if !ok {
			return nil
		}
		return upd
	}
}

func (m Model) waitForDone() tea.Cmd {
	return func() tea.Msg {
		if m.installDoneCh == nil && m.installErrCh == nil {
			return nil
		}
		select {
		case summary := <-m.installDoneCh:
			return installDoneMsg(summary)
		case err := <-m.installErrCh:
			return errMsg{err: err}
		}
	}
}

func (m Model) View() string {
	switch m.state {
	case StateWelcome:
		return welcomeView(m.width)
	case StateScanning:
		return scanningView(m)
	case StateSelection:
		return selectionView(m)
	case StateXcodeWait:
		return xcodeView(m)
	case StateInstalling:
		return installingView(m)
	case StateSummary:
		return summaryView(m)
	default:
		return "unknown state\n"
	}
}

func welcomeView(width int) string {
	msg := titleStyle.Render("Team Mac Onboarding Tool") + "\n\nPress Enter to continue\nPress q to quit\n"
	box := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder(), true).BorderForeground(oneDarkBlue).Width(min(width-4, 72))
	return box.Render(msg)
}

func scanningView(m Model) string {
	return fmt.Sprintf("\n %s Scanning system for installed packages...\n", m.spin.View())
}

func selectionView(m Model) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select packages to install"))
	b.WriteString("\n\n")

	viewHeight := m.height - 5
	if viewHeight < 1 {
		viewHeight = 1
	}

	start := m.scrollOffset
	end := start + viewHeight
	if end > len(m.listItems) {
		end = len(m.listItems)
	}
	if start > end {
		start = end
	}

	for i := start; i < end; i++ {
		item := m.listItems[i]
		isCursor := (i == m.cursor)

		cursor := "  "
		if isCursor {
			cursor = cursorStyle.Render("▸ ")
		}

		if item.isCategory && item.category != nil {
			cat := item.category
			countSel := selectedCount(m.selected, m.packages, cat.Key)

			collapseIcon := "[-]"
			if m.collapsed[cat.Key] {
				collapseIcon = "[+]"
			}

			line := fmt.Sprintf("%s%s %s [%d selected]\n", cursor, collapseIcon, strings.ToUpper(cat.Name), countSel)

			// Style category header
			line = strings.Replace(line, strings.ToUpper(cat.Name), categoryStyle.Render(strings.ToUpper(cat.Name)), 1)

			if cat.Required {
				line = strings.TrimRight(line, "\n") + dimStyle.Render(" (required)") + "\n"
			}
			b.WriteString(line)
			continue
		}

		if item.isSubCategory {
			line := fmt.Sprintf("%s   %s\n", cursor, dimStyle.Render(strings.ToUpper(item.subCategoryName)))
			b.WriteString(line)
			continue
		}

		if item.pkg != nil {
			pkg := item.pkg

			ch := "[ ]"

			isInstalled := m.installed[pkg.Name]
			isSelected := m.selected[pkg.Name]
			isRequired := pkg.Required || pkg.Category == "fonts"

			if isRequired {
				isSelected = true
			}

			if isSelected {
				ch = "[x]"
				if !isRequired {
					ch = okStyle.Render(ch)
				}
			} else if isInstalled {
				ch = "[✓]"
				ch = okStyle.Render(ch)
			}

			indent := "  "
			if pkg.SubCategory != "" {
				indent = "    "
			}

			line := fmt.Sprintf("%s%s%s %s", cursor, indent, ch, pkg.Name)

			// Add status text
			if isInstalled {
				if isSelected && !isRequired {
					line += dimStyle.Render(" (reinstall)")
				} else if isSelected && isRequired {
					line += dimStyle.Render(" (installed)")
				}
			}

			if pkg.Required {
				line += dimStyle.Render(" (required)")
			}

			b.WriteString(line + "\n")
		}
	}

	rowsRendered := end - start
	if rowsRendered < viewHeight {
		b.WriteString(strings.Repeat("\n", viewHeight-rowsRendered))
	}

	b.WriteString(dimStyle.Render("\n[↑↓] Navigate  [Space] Toggle  [a] Select section  [n] Deselect section  [Enter] Start"))
	return b.String()
}

func selectedCount(selected map[string]bool, pkgs []config.Package, cat string) int {
	n := 0
	for _, p := range pkgs {
		if p.Category != cat {
			continue
		}
		if p.Required || p.Category == "fonts" || selected[p.Name] {
			n++
		}
	}
	return n
}

func xcodeView(m Model) string {
	return titleStyle.Render("Xcode Command Line Tools") + "\n\n" +
		"macsetup needs Xcode Command Line Tools.\n" +
		"The install dialog should be open. Complete it, then wait here.\n\n" +
		"Press q to quit.\n"
}

func installingView(m Model) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Installing..."))
	b.WriteString("\n\n")

	// Section 1: Installed packages (green, sorted alphabetically)
	if len(m.installedPackages) > 0 {
		var names []string
		for name := range m.installedPackages {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			msg := m.installedPackages[name]
			line := okStyle.Render("✓ " + name)
			if msg != "" {
				line += dimStyle.Render(" (" + msg + ")")
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Section 2: Failed packages (red)
	if len(m.failedPackages) > 0 {
		for name, errMsg := range m.failedPackages {
			line := badStyle.Render("✗ " + name)
			if errMsg != "" {
				line += badStyle.Render(" (" + errMsg + ")")
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Section 3: Currently running packages
	if len(m.runningPackages) > 0 {
		for name, msg := range m.runningPackages {
			line := m.spin.View() + " " + name
			if msg != "" {
				line += dimStyle.Render(" (" + msg + ")")
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Progress bar at the bottom
	var percent float64
	if m.totalPackages > 0 {
		percent = float64(m.completedPackages) / float64(m.totalPackages)
	}
	b.WriteString(m.bar.ViewAs(percent))
	b.WriteString(fmt.Sprintf("  %d/%d packages\n\n", m.completedPackages, m.totalPackages))

	b.WriteString(dimStyle.Render("This may take a while. Press q to quit.\n"))
	return b.String()
}

func summaryView(m Model) string {
	if m.err != nil {
		return badStyle.Render("Error: "+m.err.Error()) + "\n\nPress Enter to exit.\n"
	}
	var ok, skipped, failed int
	for _, r := range m.results {
		switch r.Status {
		case installer.StatusInstalled:
			ok++
		case installer.StatusSkipped:
			skipped++
		case installer.StatusFailed:
			failed++
		}
	}
	elapsed := time.Since(m.startTime).Round(time.Second)
	if m.startTime.IsZero() {
		elapsed = 0
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Summary"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s %d installed  %s %d skipped  %s %d failed  (%s)\n\n",
		okStyle.Render("+"),
		ok,
		dimStyle.Render("✓"),
		skipped,
		badStyle.Render("!"),
		failed,
		elapsed,
	))
	if failed > 0 {
		b.WriteString(badStyle.Render("Failures:\n"))
		for _, r := range m.results {
			if r.Status != installer.StatusFailed {
				continue
			}
			b.WriteString(fmt.Sprintf("- %s: %s\n", r.Package.Name, r.Error))
		}
		b.WriteString("\n")
	}
	b.WriteString("Press Enter to exit.\n")
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
