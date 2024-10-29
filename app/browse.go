package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/mwinters0/hnjobs/config"
	"github.com/mwinters0/hnjobs/db"
	"github.com/mwinters0/hnjobs/hn"
	"github.com/mwinters0/hnjobs/regionlist"
	"github.com/mwinters0/hnjobs/sanitview"
	"github.com/mwinters0/hnjobs/scoring"
	"github.com/mwinters0/hnjobs/theme"
	"github.com/rivo/tview"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// -- UI elements

// DisplayJob is a db.Job formatted for display
type DisplayJob struct {
	*db.Job
	DisplayCompany string // what's displayed in the list, e.g. "* [7] McDonald's"
	DisplayText    string // job text formatted for terminal
	Hidden         bool
}

var displayJobs []*DisplayJob

var screenSize = struct {
	X int
	Y int
}{}

var tvApp *tview.Application
var tcellScreen tcell.Screen
var companyList *tview.List
var headerText *tview.TextView
var jobText *tview.TextView
var pages *tview.Pages
var showingModal bool
var pageScrollAmount int = 10
var prevSelectedJob = -1 //for listNavHandler()

var displayStats = struct {
	numTotal          int //redundant but separates concerns
	numBelowThreshold int
	numUninterested   int
	numHidden         int
}{
	0, 0, 0, 0,
}

type DisplayStory struct {
	*hn.Story
	DisplayTitle string
}

var displayOptions = struct {
	threshold          int
	showBelowThreshold bool
	showUninterested   bool
	curStory           *DisplayStory
	urlFootnotes       bool
}{
	threshold:          1,
	showBelowThreshold: false,
	showUninterested:   false,
	curStory:           &DisplayStory{Story: &hn.Story{Id: 0}},
	urlFootnotes:       false,
}

// TODO make me responsive
const maxCompanyNameDisplayLength = 20

var curTheme theme.Theme

func reset() {
	// TODO cleanup
	displayOptions = struct {
		threshold          int
		showBelowThreshold bool
		showUninterested   bool
		curStory           *DisplayStory
		urlFootnotes       bool
	}{
		threshold:          1,
		showBelowThreshold: false,
		showUninterested:   false,
		curStory:           &DisplayStory{Story: &hn.Story{Id: 0}},
		urlFootnotes:       false,
	}
	displayStats = struct {
		numTotal          int
		numBelowThreshold int
		numUninterested   int
		numHidden         int
	}{
		0, 0, 0, 0,
	}
	displayJobs = []*DisplayJob{}
	showingModal = false
	companyList.Clear()
	headerText.Clear()
	jobText.Clear()
	prevSelectedJob = -1
}

// -- misc

func maybePanic(err error) {
	if err != nil {
		tvApp.Suspend(func() {
			panic(err)
		})
	}
}

var preRegex *regexp.Regexp
var linkRegex *regexp.Regexp
var emailRegex *regexp.Regexp

// formats the job for TTY
func newDisplayJob(job *db.Job) *DisplayJob {
	var err error
	dj := &DisplayJob{
		Job:            job,
		DisplayCompany: formatDisplayCompany(job),
	}
	if job.Score < displayOptions.threshold && !displayOptions.showBelowThreshold {
		dj.Hidden = true
	}
	if !job.Interested && !displayOptions.showUninterested {
		dj.Hidden = true
	}

	str := job.Text
	str = html.UnescapeString(str)
	str = strings.ReplaceAll(str, "<p>", "\n\n")
	str = strings.ReplaceAll(str, "<code>", "") // We always have <pre><code> so don't need both
	str = strings.ReplaceAll(str, "</code>", "")
	str = tview.Escape(str)
	str = fmt.Sprintf( // as html so it can get formatted later along with other links
		"%s\n\n\n%s►%s Original HN comment: <a href=\"https://news.ycombinator.com/item?id=%d\"> </a>",
		str,
		curTheme.JobBody.CompanyName.AsTag(),
		curTheme.JobBody.Normal.AsTag(),
		job.Id,
	)

	const (
		keyFg int = iota
		keyBg
		keyAttr
		keyURL
	)
	rlm := regionlist.NewRegionListManager(len(str))
	err = rlm.CreateRegionList(keyFg, curTheme.JobBody.Normal.Fg)
	if err != nil {
		panic(err)
	}
	err = rlm.CreateRegionList(keyBg, curTheme.JobBody.Normal.Bg)
	if err != nil {
		panic(err)
	}
	err = rlm.CreateRegionList(keyAttr, "-")
	if err != nil {
		panic(err)
	}
	err = rlm.CreateRegionList(keyURL, "-")
	if err != nil {
		panic(err)
	}

	// helpers
	addStyleRegion := func(start int, end int, style *sanitview.TViewStyle) {
		if style.Fg != "" {
			err = rlm.InsertRegion(keyFg, &regionlist.Region{
				Start: start,
				End:   end,
				Value: style.Fg,
			})
			if err != nil {
				panic(err)
			}
		}
		if style.Bg != "" {
			err = rlm.InsertRegion(keyBg, &regionlist.Region{
				Start: start,
				End:   end,
				Value: style.Bg,
			})
			if err != nil {
				panic(err)
			}
		}
		if style.Attrs != "" {
			err = rlm.InsertRegion(keyAttr, &regionlist.Region{
				Start: start,
				End:   end,
				Value: style.Attrs,
			})
			if err != nil {
				panic(err)
			}
		}
		if style.Url != "" {
			err = rlm.InsertRegion(keyURL, &regionlist.Region{
				Start: start,
				End:   end,
				Value: style.Url,
			})
			if err != nil {
				panic(err)
			}
		}
	}
	processSubmatchRegex := func(r *regexp.Regexp, s string, rlm *regionlist.RegionListManager, style *sanitview.TViewStyle) string {
		offset := 0
		matches := r.FindAllStringSubmatchIndex(s, -1)
		for _, m := range matches {
			matchStyle := style.Clone()
			matchStart := m[0]
			matchEnd := m[1]
			matchLen := matchEnd - matchStart
			submatchStart := m[2]
			submatchEnd := m[3]
			submatchLen := submatchEnd - submatchStart
			submatchVal := str[submatchStart:submatchEnd]
			if submatchLen == 0 {
				continue
			}
			editSizeDelta := submatchLen - matchLen
			s = s[:matchStart+offset] + submatchVal + s[matchEnd+offset:]
			_, err = rlm.ResizeAt(matchStart+offset, editSizeDelta)
			if err != nil {
				panic(err)
			}
			if matchStyle.Url == "SUBMATCH" {
				matchStyle.Url = submatchVal
			}
			addStyleRegion(matchStart+offset, matchStart+offset+submatchLen, matchStyle)
			offset += editSizeDelta
		}
		return s
	}

	// company name
	addStyleRegion(0, len(job.Company), curTheme.JobBody.CompanyName)

	//pre
	if preRegex == nil {
		preRegex = regexp.MustCompile(`(?s)<pre>(?P<content>.+?)</pre>`)
	}
	str = processSubmatchRegex(preRegex, str, rlm, curTheme.JobBody.Pre)

	// url
	if linkRegex == nil {
		linkRegex = regexp.MustCompile(`(?U)<a href="(?P<href>.+)".*</a>`)
	}
	urlStyle := sanitview.MergeTviewStyles(curTheme.JobBody.URL, &sanitview.TViewStyle{Url: "SUBMATCH"})
	str = processSubmatchRegex(linkRegex, str, rlm, urlStyle)

	// email
	if emailRegex == nil {
		emailRegex = regexp.MustCompile(`(?P<email>[a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\.[a-zA-Z0-9_-]+)`)
	}
	str = processSubmatchRegex(emailRegex, str, rlm, curTheme.JobBody.Email)

	// rules
	for _, r := range scoring.GetRules() {
		switch r.RuleType {
		case scoring.TextFound:
			matchIndices := r.Regex.FindAllStringIndex(str, -1)
			if matchIndices != nil {
				var style *sanitview.TViewStyle
				if r.Score >= 0 {
					style = curTheme.JobBody.PositiveHit
				} else {
					style = curTheme.JobBody.NegativeHit
				}
				for _, pair := range matchIndices {
					addStyleRegion(pair[0], pair[1]-1, style)
				}
			}
		}
	}

	// apply regions
	offset := 0
	for e := range rlm.MergedEvents() {
		fg := e.Values[keyFg]
		bg := e.Values[keyBg]
		attr := e.Values[keyAttr]
		url := e.Values[keyURL]
		tag := fmt.Sprintf("[%s:%s:%s:%s]", fg, bg, attr, url)
		tagLen := len(tag)
		totalOffset := e.Offset + offset
		str = str[:totalOffset] + tag + str[totalOffset:]
		offset += tagLen
	}

	dj.DisplayText = str
	return dj
}

func formatDisplayCompany(job *db.Job) string {
	//truncate company name for display in list
	cname := job.Company
	if len(cname) > maxCompanyNameDisplayLength {
		scoreLen := len(strconv.Itoa(job.Score)) - 1 // handle large scores - assume size of 1
		cname = cname[:maxCompanyNameDisplayLength-3-scoreLen] + "..."
	}

	cl := curTheme.CompanyList

	statusChar := cl.Chars.Read
	statusCharStyle := cl.Colors.StatusChar
	scoreStyle := cl.Colors.Score
	nameStyle := cl.Colors.CompanyName
	if !job.Interested {
		statusCharStyle = sanitview.MergeTviewStyles(statusCharStyle, cl.Colors.StatusCharUninterested)
		scoreStyle = sanitview.MergeTviewStyles(scoreStyle, cl.Colors.ScoreUninterested)
		nameStyle = sanitview.MergeTviewStyles(nameStyle, cl.Colors.CompanyNameUninterested)
		statusChar = cl.Chars.Uninterested
	} else {
		if !job.Read {
			statusCharStyle = sanitview.MergeTviewStyles(statusCharStyle, cl.Colors.StatusCharUnread)
			scoreStyle = sanitview.MergeTviewStyles(scoreStyle, cl.Colors.ScoreUnread)
			nameStyle = sanitview.MergeTviewStyles(nameStyle, cl.Colors.CompanyNameUnread)
			statusChar = cl.Chars.Unread
		}
		if job.Priority {
			statusCharStyle = sanitview.MergeTviewStyles(statusCharStyle, cl.Colors.StatusCharPriority)
			scoreStyle = sanitview.MergeTviewStyles(scoreStyle, cl.Colors.ScorePriority)
			nameStyle = sanitview.MergeTviewStyles(nameStyle, cl.Colors.CompanyNamePriority)
			statusChar = cl.Chars.Priority
		}
		if job.Applied {
			statusCharStyle = sanitview.MergeTviewStyles(statusCharStyle, cl.Colors.StatusCharApplied)
			scoreStyle = sanitview.MergeTviewStyles(scoreStyle, cl.Colors.ScoreApplied)
			nameStyle = sanitview.MergeTviewStyles(nameStyle, cl.Colors.CompanyNameApplied)
			statusChar = cl.Chars.Applied
		}
	}
	statusCharStyleTag := sanitview.StyleToString(statusCharStyle)
	scoreStyleTag := sanitview.StyleToString(scoreStyle)
	nameStyleTag := sanitview.StyleToString(nameStyle)

	score := tview.Escape(fmt.Sprintf("[%d]", job.Score))
	cname = fmt.Sprintf(
		"%s%s %s%s %s%s",
		statusCharStyleTag, statusChar,
		scoreStyleTag, score,
		nameStyleTag, cname,
	)
	return cname
}

func newDisplayStory(s *hn.Story) *DisplayStory {
	ds := &DisplayStory{
		Story:        s,
		DisplayTitle: s.Title[8:], // chop off "Ask HN: "
	}
	return ds
}

func setupLatestStory() {
	if displayOptions.curStory.Id == 0 {
		// get latest
		latest, err := db.GetLatestStory()
		if errors.Is(err, db.ErrNoResults) {
			// probably first run
			return
		}
		if err != nil {
			panic(fmt.Errorf("error finding latest job story from DB: %v", err))
		}
		displayOptions.curStory = newDisplayStory(latest)
	}
}

func Browse() {
	var err error
	displayOptions.threshold = config.GetConfig().Display.ScoreThreshold
	curTheme = theme.GetTheme()
	setupLatestStory()

	tview.Borders = struct {
		Horizontal  rune
		Vertical    rune
		TopLeft     rune
		TopRight    rune
		BottomLeft  rune
		BottomRight rune

		LeftT   rune
		RightT  rune
		TopT    rune
		BottomT rune
		Cross   rune

		HorizontalFocus  rune
		VerticalFocus    rune
		TopLeftFocus     rune
		TopRightFocus    rune
		BottomLeftFocus  rune
		BottomRightFocus rune
	}{
		Horizontal:  ' ', // no borders on unfocused elements
		Vertical:    ' ',
		TopLeft:     ' ',
		TopRight:    ' ',
		BottomLeft:  ' ',
		BottomRight: ' ',

		LeftT:   tview.BoxDrawingsLightVerticalAndRight,
		RightT:  tview.BoxDrawingsLightVerticalAndLeft,
		TopT:    tview.BoxDrawingsLightDownAndHorizontal,
		BottomT: tview.BoxDrawingsLightUpAndHorizontal,
		Cross:   tview.BoxDrawingsLightVerticalAndHorizontal,

		HorizontalFocus:  tview.BoxDrawingsLightHorizontal,
		VerticalFocus:    tview.BoxDrawingsLightVertical,
		TopLeftFocus:     tview.BoxDrawingsLightDownAndRight,
		TopRightFocus:    tview.BoxDrawingsLightDownAndLeft,
		BottomLeftFocus:  tview.BoxDrawingsLightUpAndRight,
		BottomRightFocus: tview.BoxDrawingsLightUpAndLeft,
	}

	// don't want these on the tvApp because then they fire even when modals are up
	sharedRuneHandler := func(r rune) (handled bool) {
		switch r {
		case 'a':
			if !showingModal {
				actionListMarkApplied()
			}
			return true
		case 'f':
			if !showingModal {
				actionConsiderFetch(false)
			}
			return true
		case 'F':
			if !showingModal {
				actionConsiderFetch(true)
			}
			return true
		case 'm':
			if !showingModal {
				actionBrowseStories()
			}
			return true
		case 'p':
			if !showingModal {
				actionListMarkPriority()
			}
			return true
		case 'r':
			if !showingModal {
				actionListMarkRead()
			}
			return true
		case 's':
			if !showingModal {
				actionRescore()
			}
			return true
		case 'T':
			if !showingModal {
				actionToggleShowBelowThreshold()
			}
			return true
		case 'x':
			if !showingModal {
				actionListMarkInterested()
			}
			return true
		case 'X':
			if !showingModal {
				actionToggleShowUninterested()
			}
			return true
		}
		return false
	}

	tvApp = tview.NewApplication()
	tvApp.EnableMouse(true)
	tvApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			tvApp.Stop()
		}
		return event
	})

	focusColor := tcell.GetColor(curTheme.UI.FocusBorder.Fg)
	// list
	listBg := tcell.GetColor(curTheme.CompanyList.Colors.FrameBackground.Bg)
	companyList = tview.NewList(). // list attrs
					ShowSecondaryText(false).
					SetWrapAround(false).
					SetChangedFunc(listNavHandler).
					SetSelectedBackgroundColor(tcell.GetColor(curTheme.CompanyList.Colors.SelectedItemBackground.Bg))
	companyList. // Box attrs
			SetBackgroundColor(listBg)
	companyList.SetHighlightFullLine(true)
	companyList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if sharedRuneHandler(event.Rune()) {
			return nil
		}
		switch event.Rune() {
		case 'g':
			companyList.SetCurrentItem(companyList.GetItemCount())
			return nil
		case 'G':
			companyList.SetCurrentItem(0)
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		switch event.Key() {
		case tcell.KeyTab:
			tvApp.SetFocus(jobText)
			return nil
		case tcell.KeyCtrlD:
			if len(displayJobs) == 0 {
				return nil
			}
			c := companyList.GetCurrentItem()
			target := c + pageScrollAmount
			if target < len(displayJobs)-1 {
				// page down
				companyList.SetCurrentItem(target)
			} else {
				// select last
				companyList.SetCurrentItem(len(displayJobs) - 1)
			}
			return nil
		case tcell.KeyCtrlU:
			if len(displayJobs) == 0 {
				return nil
			}
			c := companyList.GetCurrentItem()
			target := c - pageScrollAmount
			if target > 0 {
				// page up
				companyList.SetCurrentItem(c - pageScrollAmount)
			} else {
				// select first
				companyList.SetCurrentItem(0)
			}
			return nil
		case tcell.KeyF1:
			if !showingModal {
				actionHelp()
			}
			return nil
		}
		return event
	})
	listFrameInner := tview.NewFrame(companyList)
	listFrameInner.SetBackgroundColor(listBg)
	clFrameTransitionStyle := &sanitview.TViewStyle{
		Fg: curTheme.CompanyList.Colors.FrameHeader.Bg,
		Bg: curTheme.CompanyList.Colors.FrameBackground.Bg,
	}
	listFrame := tview.NewFrame(listFrameInner).
		AddText(
			fmt.Sprintf(
				"%s%s%s%s%s%s",
				clFrameTransitionStyle.AsTag(),
				"◢",
				curTheme.CompanyList.Colors.FrameHeader.AsTag(),
				" Company ",
				clFrameTransitionStyle.AsTag(),
				"◤",
			),
			true, tview.AlignCenter, 0,
		).
		SetBorders(0, 0, 0, 0, 0, 0)
	listFrame. // box attrs
			SetBackgroundColor(listBg).
			SetBorder(true).
			SetBorderColor(focusColor)

	// header
	headerText = tview.NewTextView(). // tv attrs
						SetDynamicColors(true).
						SetScrollable(false)
	headerText. //box attrs
			SetBackgroundColor(tcell.GetColor(curTheme.UI.HeaderStatsHidden.Bg))

	// job
	jobText = tview.NewTextView(). //tv attrs
					SetDynamicColors(true).
					SetRegions(true)
	jobText. // Box attrs
			SetBackgroundColor(tcell.GetColor(curTheme.JobBody.FrameBackground.Bg))
	jobText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if sharedRuneHandler(event.Rune()) {
			return nil
		}
		switch event.Key() {
		case tcell.KeyTab:
			tvApp.SetFocus(companyList)
		}
		return event
	})
	jobFrameInner := tview.NewFrame(jobText) // frame attrs
	jobFrameInner.SetBackgroundColor(tcell.GetColor(curTheme.JobBody.FrameBackground.Bg))
	jFrameTransitionStyle := &sanitview.TViewStyle{
		Fg: curTheme.JobBody.FrameHeader.Bg,
		Bg: curTheme.JobBody.FrameBackground.Bg,
	}
	jobFrame := tview.NewFrame(jobFrameInner).
		AddText(
			fmt.Sprintf(
				"%s%s%s%s%s%s",
				jFrameTransitionStyle.AsTag(),
				"◢",
				curTheme.JobBody.FrameHeader.AsTag(),
				" Job ",
				jFrameTransitionStyle.AsTag(),
				"◤",
			),
			true, tview.AlignCenter, 0,
		).
		SetBorders(0, 0, 0, 0, 0, 0)
	jobFrame. // Box attrs
			SetBackgroundColor(tcell.GetColor(curTheme.JobBody.FrameBackground.Bg)).
			SetBorder(true).
			SetBorderColor(focusColor)

	grid := tview.NewGrid().
		SetRows(1, 0).
		SetColumns(maxCompanyNameDisplayLength+11, 0)
	grid.AddItem(headerText, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(listFrame, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(jobFrame, 1, 1, 1, 1, 0, 0, false)

	pages = tview.NewPages().AddPage("base", grid, true, true)

	tcellScreen, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	tvApp.SetScreen(tcellScreen)
	tvApp.SetBeforeDrawFunc(bdf)
	tvApp.SetRoot(pages, true)

	if displayOptions.curStory.Id == 0 {
		// we couldn't find a latest story in the DB
		rebuildHeaderText()
		actionConsiderFetch(false)
	} else {
		loadList(0)
	}

	if err = tvApp.Run(); err != nil {
		panic(err)
	}
}

func bdf(s tcell.Screen) bool {
	screenSize.X, screenSize.Y = tcellScreen.Size()
	pageScrollAmount = (screenSize.Y - 6) / 2
	rebuildHeaderText()
	return false
}

var inlineStyleBgRegex *regexp.Regexp

func replaceInlineStyleBgs(s string, bg string) string {
	bgLen := len(bg)
	if inlineStyleBgRegex == nil {
		inlineStyleBgRegex = regexp.MustCompile(`\[.+?:(?P<bg>.+?)[:\]]`)
	}
	offset := 0
	matches := inlineStyleBgRegex.FindAllStringSubmatchIndex(s, -1)
	for _, m := range matches {
		// replace the inline bg style tags
		submatchStart := m[2]
		submatchEnd := m[3]
		submatchLen := m[3] - m[2]
		if submatchLen == 0 {
			continue
		}
		diff := bgLen - submatchLen
		s = s[:submatchStart+offset] + bg + s[submatchEnd+offset:]
		offset += diff
	}
	return s
}

func fixItemBg(index int) {
	// tview doesn't do cascading styles, so when our list items have inline styles they override the
	// "selected item" background color.  We then need to manually put the "selected" bg back into those items.
	// We choose to leave the Priority jobs with "broken" backgrounds.
	if !(displayJobs[index].Priority) {
		s := replaceInlineStyleBgs(
			displayJobs[index].DisplayCompany,
			curTheme.CompanyList.Colors.SelectedItemBackground.Bg,
		)
		companyList.SetItemText(index, s, "")
	}
}

func listNavHandler(index int, mainText string, secondaryText string, shortcut rune) {
	if companyList.GetItemCount() < len(displayJobs) {
		// loadList is currently building the list
		return
	}
	jobText.SetText(displayJobs[index].DisplayText)

	fixItemBg(index)
	if prevSelectedJob != -1 {
		// restore the previously-selected item back to unmodified
		companyList.SetItemText(prevSelectedJob, displayJobs[prevSelectedJob].DisplayCompany, "")
	}
	prevSelectedJob = index
}

// listItemModified is called after we modify an item in the list. It persists the modification and updates the display.
func listItemModified(i int) error {
	displayJobs[i].DisplayCompany = formatDisplayCompany(displayJobs[i].Job)
	displayJobs[i].Job.ReviewedTime = time.Now().UTC().Unix()
	err := db.UpsertJob(displayJobs[i].Job)
	maybePanic(err)
	fixItemBg(i)
	companyList.SetCurrentItem(i) //triggers redraw
	rebuildHeaderText()
	return nil
}

func makeModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false),
			width, 1, true,
		).
		AddItem(nil, 0, 1, false)
}

func showModalTextView(rows int, cols int, text string, title string) {
	// I hoped to get more mileage out of this function than I have so far ...
	if !showingModal {
		showingModal = true
		bgColor := tcell.GetColor(curTheme.UI.ModalNormal.Bg)
		modalTV := tview.NewTextView(). //TextView attrs
						SetSize(rows, cols).
						SetDynamicColors(true).
						SetTextStyle(curTheme.UI.ModalNormal.AsTCellStyle()) // sets modal bg
		modalTV.SetBorder(true). //Box attrs
						SetBorderColor(tcell.GetColor(curTheme.UI.FocusBorder.Fg)).
						SetTitle(curTheme.UI.ModalTitle.AsTag() + title).
						SetTitleAlign(tview.AlignLeft).
						SetBackgroundColor(bgColor)
		modalTV.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				pages.RemovePage("modalText")
				showingModal = false
				return nil
			}
			return event
		})
		modalTV.SetText(text)

		pages.AddPage("modalText", makeModal(modalTV, cols, rows+2), true, true)
		pages.SetBackgroundColor(bgColor)

		tvApp.SetFocus(modalTV)
	}
}

func actionHelp() {
	url := "https://github.com/mwinters0/hnjobs"
	hl := curTheme.UI.ModalHighlight.AsTag()
	normal := curTheme.UI.ModalNormal.AsTag()
	link := sanitview.MergeTviewStyles(
		curTheme.UI.ModalHighlight,
		&sanitview.TViewStyle{Url: url},
	).AsTag()
	helpText := normal + `
 Bindings:
   - Basics
     - ` + hl + `ESC` + normal + ` - close dialogs
     - ` + hl + `TAB` + normal + ` - switch focus
     - ` + hl + `jk` + normal + ` and ` + hl + `up/down` + normal + ` - scroll
     - ` + hl + `f` + normal + ` - fetch latest (respecting TTL)
     - ` + hl + `F` + normal + ` - force fetch all
     - ` + hl + `q` + normal + ` - quit
   - Job Filtering
     - ` + hl + `r` + normal + ` - mark read / unread
     - ` + hl + `x` + normal + ` - mark job uninterested (hidden) or interested (default)
     - ` + hl + `p` + normal + ` - mark priority / not priority
     - ` + hl + `a` + normal + ` - mark applied to / not applied to
   - Display
     - ` + hl + `X` + normal + ` - toggle hiding jobs marked uninterested
     - ` + hl + `T` + normal + ` - toggle hiding jobs below score threshold
   - Misc
     - ` + hl + `m` + normal + ` - select month (if multiple in DB) / delete old data
     - ` + hl + `s` + normal + ` - reload scoring config and re-score the jobs

 For more info: ` + link + url + normal + `
`
	showModalTextView(25, 70, helpText, " Help ")
}

func actionListMarkPriority() {
	if len(displayJobs) == 0 {
		return
	}
	i := companyList.GetCurrentItem()
	displayJobs[i].Priority = !displayJobs[i].Priority
	err := listItemModified(i)
	maybePanic(err)
}

func actionListMarkRead() {
	if len(displayJobs) == 0 {
		return
	}
	i := companyList.GetCurrentItem()
	displayJobs[i].Read = !displayJobs[i].Read
	err := listItemModified(i)
	maybePanic(err)
}

func actionListMarkInterested() {
	if len(displayJobs) == 0 {
		return
	}
	i := companyList.GetCurrentItem()
	if displayJobs[i].Interested {
		displayJobs[i].Interested = false
		displayStats.numUninterested++
	} else {
		displayJobs[i].Interested = true
		displayStats.numUninterested--
	}
	err := listItemModified(i)
	maybePanic(err)
}

func actionListMarkApplied() {
	if len(displayJobs) == 0 {
		return
	}
	i := companyList.GetCurrentItem()
	displayJobs[i].Applied = !displayJobs[i].Applied
	err := listItemModified(i)
	maybePanic(err)
}

func actionToggleShowBelowThreshold() {
	if len(displayJobs) == 0 {
		return
	}
	displayOptions.showBelowThreshold = !displayOptions.showBelowThreshold
	curSelectedJobId := displayJobs[companyList.GetCurrentItem()].Id
	loadList(curSelectedJobId)
}

func actionToggleShowUninterested() {
	if len(displayJobs) == 0 {
		return
	}
	displayOptions.showUninterested = !displayOptions.showUninterested
	curSelectedJobId := displayJobs[companyList.GetCurrentItem()].Id
	loadList(curSelectedJobId)
}

func actionConsiderFetch(force bool) {
	if displayOptions.curStory.Id == 0 {
		// no story loaded, e.g. first launch with empty database
		fetchJobs(false)
		return
	}
	fetchJobs(force)
}

func fetchJobs(force bool) {
	showingModal = true
	ctx, cancelFetch := context.WithCancel(context.Background())
	done := false
	pDone := &done
	var rows, cols int
	if screenSize.Y > 25 {
		rows = screenSize.Y - 20
	} else {
		rows = 20
	}
	if screenSize.X > 80 {
		cols = screenSize.X - 10
		if cols > 100 {
			cols = 100
		}
	} else {
		cols = 70
	}
	bgColor := tcell.GetColor(curTheme.UI.ModalNormal.Bg)
	fetchText := tview.NewTextView().
		SetSize(rows, cols).
		SetDynamicColors(true).
		SetTextStyle(curTheme.UI.ModalNormal.AsTCellStyle())
	fetchText.SetBorder(true).
		SetBorderColor(tcell.GetColor(curTheme.UI.FocusBorder.Fg)).
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(bgColor)
	if force {
		fetchText.SetTitle(curTheme.UI.ModalTitle.AsTag() + " Force Fetching jobs - press ESC to cancel ")
	} else {
		fetchText.SetTitle(curTheme.UI.ModalTitle.AsTag() + " Fetching jobs - press ESC to cancel ")
	}
	_, _ = fetchText.Write([]byte(curTheme.UI.ModalNormal.AsTag()))
	finish := func() {
		*pDone = true
		tvApp.QueueUpdateDraw(func() {
			fetchText.SetTitle(curTheme.UI.ModalTitle.AsTag() + " Done fetching - ESC to close or up/down to review ")
			_, _ = fetchText.Write([]byte(curTheme.UI.ModalNormal.AsTag() + "\n\n Press ESC to close"))
			tvApp.SetFocus(fetchText)
		})
	}
	fetchText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			// first press = stop if WIP
			// 2nd = exit
			if !done {
				cancelFetch()
				finish()
				return nil
			}
			pages.ShowPage("base")
			pages.RemovePage("fetchModal")
			if displayOptions.curStory.Id == 0 {
				setupLatestStory()
			}
			curStory := displayOptions.curStory
			reset()
			displayOptions.curStory = curStory
			if len(displayJobs) > 0 {
				curSelectedJobId := displayJobs[companyList.GetCurrentItem()].Id
				loadList(curSelectedJobId)
			} else {
				loadList(0)
			}
			return nil
		}
		return event
	})
	addLine := func(line string) {
		if !*pDone {
			_, _ = fetchText.Write([]byte(line + "\n"))
			fetchText.ScrollToEnd()
		}
	}

	pages.AddPage("fetchModal", makeModal(fetchText, cols, rows+2), true, true)
	pages.HidePage("base")
	tvApp.SetFocus(fetchText)

	updateCallback := func(fsu FetchStatusUpdate) {
		lineStyle := curTheme.JobBody.Normal
		switch fsu.UpdateType {
		case UpdateTypeFatal, UpdateTypeNonFatalErr:
			lineStyle = curTheme.JobBody.NegativeHit
		case UpdateTypeBadComment:
			lineStyle = curTheme.JobBody.Pre
		case UpdateTypeJobFetched:
			if fsu.Value >= displayOptions.threshold {
				lineStyle = curTheme.JobBody.PositiveHit
			}
		case UpdateTypeNewStory:
			lineStyle = curTheme.JobBody.PositiveHit
		case UpdateTypeDone:
			addLine(" --------------------\n")
			if fsu.Value > 0 {
				// new jobs fetched
				lineStyle = curTheme.UI.HeaderStatsNormal
			}
		}
		tvApp.QueueUpdateDraw(func() {
			addLine(" " + lineStyle.AsTag() + tview.Escape(fsu.Message))
		})
	}
	doneCallback := func(err error) {
		finish()
	}
	status := make(chan FetchStatusUpdate)
	fo := FetchOptions{
		Context:     ctx,
		Status:      status,
		ModeForce:   force,
		StoryID:     0,
		TTLSec:      config.GetConfig().Cache.TTLSecs,
		MustContain: WhoIsHiringString,
	}
	go FetchAsync(fo)
	go func() {
		for {
			select {
			case fsu, ok := <-status:
				if !ok {
					//EOF
					doneCallback(nil)
					return
				}
				updateCallback(fsu)
				if fsu.Error != nil {
					if fsu.UpdateType == UpdateTypeFatal {
						doneCallback(fsu.Error)
						return
					}
				}
			}
		}
	}()
}

func actionBrowseStories() {
	var err error
	var stories []*DisplayStory
	rows := 20
	cols := 60
	textviewRows := 5
	_ = textviewRows
	if !showingModal {
		showingModal = true
		bgColor := tcell.GetColor(curTheme.UI.ModalNormal.Bg)
		storyListItemStyle := curTheme.UI.ModalHighlight
		const browseStoriesPageName = "browseStories"

		const deleteConfirmPageName = "storyDeleteConfirm"
		var deleteStory func() // defined after storyList
		deleteConfirmModal := tview.NewModal().
			AddButtons([]string{"Delete", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				switch buttonLabel {
				case "Delete":
					deleteStory()
				case "Cancel":
					pages.RemovePage(deleteConfirmPageName)
				}
			}).
			SetBackgroundColor(bgColor)
		deleteConfirmModal. //box attrs
					SetBorderColor(tcell.GetColor(curTheme.UI.FocusBorder.Fg)).
					SetBorderStyle(curTheme.UI.FocusBorder.AsTCellStyle())
		deleteConfirmModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				pages.RemovePage(deleteConfirmPageName)
				return nil
			}
			return event
		})

		storyList := tview.NewList(). // list attrs
						ShowSecondaryText(false).
						SetWrapAround(false).
						SetSelectedBackgroundColor(tcell.GetColor(curTheme.CompanyList.Colors.SelectedItemBackground.Bg))
		storyList. // Box attrs
				SetBackgroundColor(bgColor)
		storyList.SetHighlightFullLine(true)
		storyList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'd':
				if storyList.GetItemCount() > 0 {
					i := storyList.GetCurrentItem()
					s := fmt.Sprintf("Delete %s%s%s?", storyListItemStyle.AsTag(), stories[i].DisplayTitle, curTheme.UI.ModalNormal.AsTag())
					deleteConfirmModal.SetText(s)
					pages.AddPage(deleteConfirmPageName, makeModal(deleteConfirmModal, 30, 5), true, true)
					tvApp.SetFocus(deleteConfirmModal)
				}
				return nil
			case 'g':
				storyList.SetCurrentItem(storyList.GetItemCount())
				return nil
			case 'G':
				storyList.SetCurrentItem(0)
				return nil
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}
			switch event.Key() {
			case tcell.KeyEscape:
				pages.RemovePage(browseStoriesPageName)
				showingModal = false
				return nil
			}
			return event
		})
		getStoryTitle := func(s *DisplayStory, storyStyle *sanitview.TViewStyle) string {
			return fmt.Sprintf("%s[ %v ]", storyStyle.AsTag(), s.DisplayTitle) //strip "Ask HN: "
		}
		prevSelectedStory := -1
		storyListNavHandler := func(index int, mainText string, secondaryText string, shortcut rune) {
			if len(stories) == 0 {
				return
			}
			// fix selected bgs
			s := replaceInlineStyleBgs(
				getStoryTitle(stories[index], storyListItemStyle),
				curTheme.CompanyList.Colors.SelectedItemBackground.Bg,
			)
			storyList.SetItemText(index, s, "")
			if prevSelectedStory != -1 {
				storyList.SetItemText(prevSelectedStory, getStoryTitle(stories[prevSelectedStory], storyListItemStyle), "")
			}
			prevSelectedStory = index
		}
		storyList.SetChangedFunc(storyListNavHandler)
		storyList.SetSelectedFunc(func(i int, s string, s2 string, r rune) {
			pages.RemovePage(browseStoriesPageName)
			if stories[i].Id == displayOptions.curStory.Id {
				return
			}
			reset()
			displayOptions.curStory = stories[i]
			loadList(0)
		})
		curStoryIndex := -1
		loadStoryList := func() {
			// called here and after delete
			dbStories, err := db.GetAllStories()
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				tvApp.Suspend(func() {
					panic(err)
				})
			}
			storyList.Clear()
			stories = []*DisplayStory{}
			for i, story := range dbStories {
				displayStory := newDisplayStory(story)
				stories = append(stories, displayStory)
				if story.Id == displayOptions.curStory.Id {
					curStoryIndex = i
				}
				storyList.AddItem(getStoryTitle(displayStory, storyListItemStyle), "", 0, nil)
			}
		}
		loadStoryList()
		storyListFrameInner := tview.NewFrame(storyList) //frame attrs
		storyListFrameInner.SetBackgroundColor(bgColor)
		storyListFrame := tview.NewFrame(storyListFrameInner). //frame attrs
									AddText(
				"  enter to select, d to delete  ",
				true, tview.AlignCenter, 0,
			).
			SetBorders(1, 0, 0, 0, 0, 0)
		storyListFrame. // box attrs
				SetBorder(true).
				SetBorderColor(tcell.GetColor(curTheme.UI.FocusBorder.Fg)).
				SetBackgroundColor(bgColor).
				SetTitle(curTheme.UI.ModalTitle.AsTag() + " Select Month ").
				SetTitleAlign(tview.AlignLeft)

		storyGrid := tview.NewGrid(). //grid attrs
						SetRows(rows).
						SetColumns(cols)
		storyGrid.AddItem(storyListFrame, 0, 0, 1, 1, 0, 0, true)

		deleteStory = func() {
			i := storyList.GetCurrentItem()
			storyId := stories[i].Id
			prevSelectedStory = -1
			err = db.DeleteStoryAndJobsByStoryID(storyId)
			maybePanic(err)
			if storyId == displayOptions.curStory.Id {
				reset()
				// todo? try to auto-load a different story
			}
			loadStoryList()
			pages.RemovePage(deleteConfirmPageName)
		}

		pages.AddPage(browseStoriesPageName, makeModal(storyGrid, cols, rows), true, true)

		if curStoryIndex != -1 {
			storyList.SetCurrentItem(curStoryIndex)
		} else {
			storyList.SetCurrentItem(0)
			listNavHandler(0, "", "", 0)
		}
		tvApp.SetFocus(storyList)
	}
}

func actionRescore() {
	if len(displayJobs) == 0 {
		return
	}
	var err error
	rows := 5
	cols := 50
	const pageName string = "rescoreText"
	if !showingModal {
		showingModal = true
		bgColor := tcell.GetColor(curTheme.UI.ModalNormal.Bg)
		modalTV := tview.NewTextView(). //TextView attrs
						SetSize(rows, cols).
						SetDynamicColors(true).
						SetTextStyle(curTheme.UI.ModalNormal.AsTCellStyle()) // sets modal bg
		modalTV.SetBorder(true). //Box attrs
						SetBorderColor(tcell.GetColor(curTheme.UI.FocusBorder.Fg)).
						SetTitle(curTheme.UI.ModalTitle.AsTag() + " Rescore ").
						SetTitleAlign(tview.AlignLeft).
						SetBackgroundColor(bgColor)
		modalTV.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				pages.RemovePage(pageName)
				showingModal = false
				return nil
			}
			return event
		})

		modalTV.SetText(fmt.Sprintf("\n Rescoring \"%s\" ...\n", displayOptions.curStory.DisplayTitle))

		pages.AddPage(pageName, makeModal(modalTV, cols, rows+2), true, true)
		pages.SetBackgroundColor(bgColor)

		tvApp.SetFocus(modalTV)

		err = config.Reload()
		maybePanic(err)
		err = scoring.ReloadRules()
		maybePanic(err)
		num, err := ReScore(displayOptions.curStory.Id)
		maybePanic(err)
		_, err = modalTV.Write([]byte(fmt.Sprintf(" Rescored %d jobs.  Press ESC to continue.", num)))
		maybePanic(err)
		story := displayOptions.curStory
		curSelectedJobId := displayJobs[companyList.GetCurrentItem()].Id
		reset()
		displayOptions.curStory = story
		loadList(curSelectedJobId)
	}

}

// loadList (re)loads dataset from DB based on displayOptions.  Assumes you've already fetched and set curStoryID.
// We try to re-select prevDisplayJobID after reload, but it may have become invisible.
func loadList(prevDisplayJobID int) {
	jobs, err := db.GetAllJobsByStoryId(displayOptions.curStory.Id, db.OrderScoreDesc)
	maybePanic(err)
	if len(jobs) == 0 {
		maybePanic(fmt.Errorf("found zero jobs in the db for curStoryID %d", displayOptions.curStory.Id))
	}
	displayStats.numTotal = len(jobs)

	companyList.Clear()
	displayJobs = []*DisplayJob{}
	displayStats.numBelowThreshold = 0
	displayStats.numUninterested = 0
	displayStats.numHidden = 0
	// rebuild list and try to find previously-selected item by id
	newDJIndex := -1
	for _, job := range jobs {
		dj := newDisplayJob(job)
		if dj.Score < displayOptions.threshold {
			displayStats.numBelowThreshold++
		}
		if !dj.Interested {
			displayStats.numUninterested++
		}
		if dj.Hidden {
			displayStats.numHidden++
			continue
		}
		if prevDisplayJobID == dj.Id {
			// we found our previously-selected item!
			newDJIndex = len(displayJobs)
		}
		displayJobs = append(displayJobs, dj)
	}
	for _, dj := range displayJobs {
		// We have to take two passes, where pass 1 is "build displayJobs" and pass 2 is
		// "add them to the list", because tview is quite buggy (or rather, has buggy
		// opinions IMO) in how it handles the listNavHandler() callback.
		companyList.AddItem(dj.DisplayCompany, "", 0, nil)
	}

	prevSelectedJob = -1 // global version
	if newDJIndex != -1 {
		// We found the DJ to select in the list.  Simply do that.
		companyList.SetCurrentItem(newDJIndex)
		if newDJIndex == 0 {
			// tview doesn't fire the handler if the selected item is 0
			listNavHandler(0, "", "", 0)
		}
		rebuildHeaderText()
		return
	}
	if len(displayJobs) > 0 {
		companyList.SetCurrentItem(0)
		// tview doesn't fire the handler if the selected item is 0
		listNavHandler(0, "", "", 0)
	}
	rebuildHeaderText()
}

func rebuildHeaderText() {
	headerText.SetText(getStatsText(screenSize.X))
}

func getStatsText(availableWidth int) string {
	if len(displayJobs) == 0 {
		return ""
	}
	condensedThreshold := 110 // TODO take two passes so we don't have to rely on a guesstimate
	condensed := availableWidth < condensedThreshold

	builder := strings.Builder{}
	builder.WriteString(curTheme.UI.HeaderStatsDate.AsTag())
	numJobsDisplayed := displayStats.numTotal - displayStats.numHidden
	if condensed {
		builder.WriteString(tview.Escape(fmt.Sprintf(" %s ",
			getShortStoryTime(displayOptions.curStory),
		)))
		builder.WriteString(fmt.Sprintf(
			"%s J:%d ",
			curTheme.UI.HeaderStatsNormal.AsTag(),
			numJobsDisplayed,
		))
		if displayStats.numHidden > 0 {
			builder.WriteString(fmt.Sprintf(
				"(%dT) ",
				displayStats.numTotal,
			))
		}
		builder.WriteString(curTheme.UI.HeaderStatsHidden.AsTag())
	} else {
		transitionStyleDateNormal := &sanitview.TViewStyle{
			Fg: curTheme.UI.HeaderStatsDate.Bg,
			Bg: curTheme.UI.HeaderStatsNormal.Bg,
		}
		builder.WriteString(tview.Escape(fmt.Sprintf(" %s %s🭬",
			displayOptions.curStory.DisplayTitle,
			transitionStyleDateNormal.AsTag(),
		)))
		builder.WriteString(fmt.Sprintf(
			"%s Jobs: %d ",
			curTheme.UI.HeaderStatsNormal.AsTag(),
			numJobsDisplayed,
		))
		if displayStats.numHidden > 0 {
			builder.WriteString(fmt.Sprintf(
				"(%d Total) ",
				displayStats.numTotal,
			))
		}
		transitionStyleNormalHidden := &sanitview.TViewStyle{
			Fg: curTheme.UI.HeaderStatsNormal.Bg,
			Bg: curTheme.UI.HeaderStatsHidden.Bg,
		}
		builder.WriteString(tview.Escape(fmt.Sprintf("%s🭬%s",
			transitionStyleNormalHidden.AsTag(),
			curTheme.UI.HeaderStatsHidden.AsTag(),
		)))
	}

	if displayStats.numTotal == 0 || displayStats.numHidden == 0 {
		return builder.String()
	}

	// details

	var btLabel, uLabel string
	if condensed {
		btLabel = "<Th"
		uLabel = "Un"
	} else {
		btLabel = " Below Threshold "
		uLabel = " Uninterested"
	}
	builder.WriteString(" (")
	moreStats := []string{}
	if displayStats.numBelowThreshold > 0 && !displayOptions.showBelowThreshold {
		btText := fmt.Sprintf("%d%s[%d]",
			displayStats.numBelowThreshold,
			btLabel,
			displayOptions.threshold,
		)
		moreStats = append(moreStats, btText)
	}
	if displayStats.numUninterested > 0 && !displayOptions.showUninterested {
		uText := fmt.Sprintf("%d%s",
			displayStats.numUninterested,
			uLabel,
		)
		moreStats = append(moreStats, uText)
	}
	if condensed {
		builder.WriteString(strings.Join(moreStats, ","))
	} else {
		builder.WriteString(strings.Join(moreStats, ", "))
	}
	builder.WriteString(")")

	return builder.String()
}

func getShortStoryTime(s *DisplayStory) string {
	monthName := ""
	switch displayOptions.curStory.GoTime.Month() {
	case 1:
		monthName = "Jan."
	case 2:
		monthName = "Feb."
	case 3:
		monthName = "Mar."
	case 4:
		monthName = "Apr."
	case 5:
		monthName = "May"
	case 6:
		monthName = "June"
	case 7:
		monthName = "Jul."
	case 8:
		monthName = "Aug."
	case 9:
		monthName = "Sep."
	case 10:
		monthName = "Oct."
	case 11:
		monthName = "Nov."
	case 12:
		monthName = "Dec."
	default:
		monthName = "uhhh"
	}
	return fmt.Sprintf("%s %s",
		monthName,
		strconv.Itoa(s.GoTime.Year()),
	)
}
