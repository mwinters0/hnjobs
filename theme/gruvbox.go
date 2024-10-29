package theme

import "github.com/mwinters0/hnjobs/sanitview"

type GruvboxIntensity = int

const (
	GruvboxIntensityHard GruvboxIntensity = iota
	GruvboxIntensityNeutral
	GruvboxIntensitySoft
)

type GruvboxMode = int

const (
	GruvboxModeDark GruvboxMode = iota
	GruvboxModeLight
)

type gruvboxSpecific struct {
	bg0            string
	bg1            string
	bg2            string
	bg3            string
	bg4            string
	fg0            string
	fg1            string
	fg2            string
	fg3            string
	fg4            string
	red            string
	green          string
	yellow         string
	blue           string
	purple         string
	aqua           string
	orange         string
	neutral_red    string
	neutral_green  string
	neutral_yellow string
	neutral_blue   string
	neutral_purple string
	neutral_aqua   string
	dark_red       string
	dark_green     string
	dark_aqua      string
	gray           string
}

func getGruvboxTheme(m GruvboxMode, i GruvboxIntensity) Theme {
	// Converted from https://github.com/ellisonleao/gruvbox.nvim/blob/main/lua/gruvbox.lua
	var gruvbox = struct {
		dark0_hard       string
		dark0            string
		dark0_soft       string
		dark1            string
		dark2            string
		dark3            string
		dark4            string
		light0_hard      string
		light0           string
		light0_soft      string
		light1           string
		light2           string
		light3           string
		light4           string
		bright_red       string
		bright_green     string
		bright_yellow    string
		bright_blue      string
		bright_purple    string
		bright_aqua      string
		bright_orange    string
		neutral_red      string
		neutral_green    string
		neutral_yellow   string
		neutral_blue     string
		neutral_purple   string
		neutral_aqua     string
		neutral_orange   string
		faded_red        string
		faded_green      string
		faded_yellow     string
		faded_blue       string
		faded_purple     string
		faded_aqua       string
		faded_orange     string
		dark_red_hard    string
		dark_red         string
		dark_red_soft    string
		light_red_hard   string
		light_red        string
		light_red_soft   string
		dark_green_hard  string
		dark_green       string
		dark_green_soft  string
		light_green_hard string
		light_green      string
		light_green_soft string
		dark_aqua_hard   string
		dark_aqua        string
		dark_aqua_soft   string
		light_aqua_hard  string
		light_aqua       string
		light_aqua_soft  string
		gray             string
	}{
		dark0_hard:       "#1d2021",
		dark0:            "#282828",
		dark0_soft:       "#32302f",
		dark1:            "#3c3836",
		dark2:            "#504945",
		dark3:            "#665c54",
		dark4:            "#7c6f64",
		light0_hard:      "#f9f5d7",
		light0:           "#fbf1c7",
		light0_soft:      "#f2e5bc",
		light1:           "#ebdbb2",
		light2:           "#d5c4a1",
		light3:           "#bdae93",
		light4:           "#a89984",
		bright_red:       "#fb4934",
		bright_green:     "#b8bb26",
		bright_yellow:    "#fabd2f",
		bright_blue:      "#83a598",
		bright_purple:    "#d3869b",
		bright_aqua:      "#8ec07c",
		bright_orange:    "#fe8019",
		neutral_red:      "#cc241d",
		neutral_green:    "#98971a",
		neutral_yellow:   "#d79921",
		neutral_blue:     "#458588",
		neutral_purple:   "#b16286",
		neutral_aqua:     "#689d6a",
		neutral_orange:   "#d65d0e",
		faded_red:        "#9d0006",
		faded_green:      "#79740e",
		faded_yellow:     "#b57614",
		faded_blue:       "#076678",
		faded_purple:     "#8f3f71",
		faded_aqua:       "#427b58",
		faded_orange:     "#af3a03",
		dark_red_hard:    "#792329",
		dark_red:         "#722529",
		dark_red_soft:    "#7b2c2f",
		light_red_hard:   "#fc9690",
		light_red:        "#fc9487",
		light_red_soft:   "#f78b7f",
		dark_green_hard:  "#5a633a",
		dark_green:       "#62693e",
		dark_green_soft:  "#686d43",
		light_green_hard: "#d3d6a5",
		light_green:      "#d5d39b",
		light_green_soft: "#cecb94",
		dark_aqua_hard:   "#3e4934",
		dark_aqua:        "#49503b",
		dark_aqua_soft:   "#525742",
		light_aqua_hard:  "#e6e9c1",
		light_aqua:       "#e8e5b5",
		light_aqua_soft:  "#e1dbac",
		gray:             "#928374",
	}

	var g gruvboxSpecific
	switch m {
	case GruvboxModeDark:
		g = gruvboxSpecific{
			bg1:            gruvbox.dark1,
			bg2:            gruvbox.dark2,
			bg3:            gruvbox.dark3,
			bg4:            gruvbox.dark4,
			fg1:            gruvbox.light1,
			fg2:            gruvbox.light2,
			fg3:            gruvbox.light3,
			fg4:            gruvbox.light4,
			red:            gruvbox.bright_red,
			green:          gruvbox.bright_green,
			yellow:         gruvbox.bright_yellow,
			blue:           gruvbox.bright_blue,
			purple:         gruvbox.bright_purple,
			aqua:           gruvbox.bright_aqua,
			orange:         gruvbox.bright_orange,
			neutral_red:    gruvbox.neutral_red,
			neutral_green:  gruvbox.neutral_green,
			neutral_yellow: gruvbox.neutral_yellow,
			neutral_blue:   gruvbox.neutral_blue,
			neutral_purple: gruvbox.neutral_purple,
			neutral_aqua:   gruvbox.neutral_aqua,
			gray:           gruvbox.gray,
		}
		switch i {
		case GruvboxIntensitySoft:
			g.bg0 = gruvbox.dark0_soft
			g.fg0 = gruvbox.light0_soft
			g.dark_red = gruvbox.dark_red_soft
			g.dark_green = gruvbox.dark_green_soft
			g.dark_aqua = gruvbox.dark_aqua_soft
		case GruvboxIntensityNeutral:
			g.bg0 = gruvbox.dark0
			g.fg0 = gruvbox.light0
			g.dark_red = gruvbox.dark_red
			g.dark_green = gruvbox.dark_green
			g.dark_aqua = gruvbox.dark_aqua
		case GruvboxIntensityHard:
			g.bg0 = gruvbox.dark0_hard
			g.fg0 = gruvbox.light0_hard
			g.dark_red = gruvbox.dark_red_hard
			g.dark_green = gruvbox.dark_green_hard
			g.dark_aqua = gruvbox.dark_aqua_hard
		}
	case GruvboxModeLight:
		g = gruvboxSpecific{
			bg1:            gruvbox.light1,
			bg2:            gruvbox.light2,
			bg3:            gruvbox.light3,
			bg4:            gruvbox.light4,
			fg1:            gruvbox.dark1,
			fg2:            gruvbox.dark2,
			fg3:            gruvbox.dark3,
			fg4:            gruvbox.dark4,
			red:            gruvbox.faded_red,
			green:          gruvbox.faded_green,
			yellow:         gruvbox.faded_yellow,
			blue:           gruvbox.faded_blue,
			purple:         gruvbox.faded_purple,
			aqua:           gruvbox.faded_aqua,
			orange:         gruvbox.faded_orange,
			neutral_red:    gruvbox.neutral_red,
			neutral_green:  gruvbox.neutral_green,
			neutral_yellow: gruvbox.neutral_yellow,
			neutral_blue:   gruvbox.neutral_blue,
			neutral_purple: gruvbox.neutral_purple,
			neutral_aqua:   gruvbox.neutral_aqua,
			gray:           gruvbox.gray,
		}
		switch i {
		case GruvboxIntensitySoft:
			g.bg0 = gruvbox.light0_soft
			g.fg0 = gruvbox.dark0_soft
			g.dark_red = gruvbox.light_red_soft
			g.dark_green = gruvbox.light_green_soft
			g.dark_aqua = gruvbox.light_aqua_soft
		case GruvboxIntensityNeutral:
			g.bg0 = gruvbox.light0
			g.fg0 = gruvbox.dark0
			g.dark_red = gruvbox.light_red
			g.dark_green = gruvbox.light_green
			g.dark_aqua = gruvbox.light_aqua
		case GruvboxIntensityHard:
			g.bg0 = gruvbox.light0_hard
			g.fg0 = gruvbox.dark0_hard
			g.dark_red = gruvbox.light_red_hard
			g.dark_green = gruvbox.light_green_hard
			g.dark_aqua = gruvbox.light_aqua_hard
		}
	}

	bg := g.bg0
	normal := &sanitview.TViewStyle{Fg: g.fg1, Bg: bg}
	dim := &sanitview.TViewStyle{Fg: g.gray}
	priority := &sanitview.TViewStyle{Fg: bg, Bg: g.neutral_yellow}
	frameHeaders := &sanitview.TViewStyle{Fg: g.fg1, Bg: g.dark_green}

	return Theme{
		Version: 1,
		UI: UIColors{
			HeaderStatsDate:   &sanitview.TViewStyle{Fg: g.fg1, Bg: g.dark_green},
			HeaderStatsNormal: &sanitview.TViewStyle{Fg: g.fg2, Bg: g.dark_aqua},
			HeaderStatsHidden: &sanitview.TViewStyle{Fg: g.fg4, Bg: g.bg1},
			FocusBorder:       &sanitview.TViewStyle{Fg: g.fg3},
			ModalTitle:        &sanitview.TViewStyle{Fg: g.orange, Bg: bg},
			ModalNormal:       normal,
			ModalHighlight:    &sanitview.TViewStyle{Fg: g.aqua, Bg: bg},
		},
		CompanyList: CompanyList{
			Colors: CompanyListColors{
				CompanyName:             &sanitview.TViewStyle{Fg: g.neutral_blue, Bg: bg},
				CompanyNameUnread:       &sanitview.TViewStyle{Attrs: "b"},
				CompanyNamePriority:     priority,
				CompanyNameApplied:      &sanitview.TViewStyle{},
				CompanyNameUninterested: dim,
				StatusChar:              normal,
				StatusCharUnread:        &sanitview.TViewStyle{Fg: g.purple, Attrs: "b"},
				StatusCharPriority:      priority,
				StatusCharApplied:       &sanitview.TViewStyle{Fg: g.green},
				StatusCharUninterested:  dim,
				Score:                   normal,
				ScoreUnread:             &sanitview.TViewStyle{Attrs: "b"},
				ScorePriority:           priority,
				ScoreApplied:            &sanitview.TViewStyle{},
				ScoreUninterested:       dim,
				SelectedItemBackground:  &sanitview.TViewStyle{Bg: g.bg1},
				FrameBackground:         &sanitview.TViewStyle{Bg: bg},
				FrameHeader:             frameHeaders,
			},
			Chars: CompanyListChars{
				Read:         " ",
				Unread:       "*",
				Applied:      "🗹",
				Uninterested: "⨯",
				Priority:     "★",
			},
		},
		JobBody: JobBodyColors{
			Normal:          normal,
			CompanyName:     &sanitview.TViewStyle{Fg: g.blue},
			URL:             &sanitview.TViewStyle{Fg: g.neutral_yellow, Attrs: "u"},
			Email:           &sanitview.TViewStyle{Fg: g.aqua},
			PositiveHit:     &sanitview.TViewStyle{Fg: g.green},
			NegativeHit:     &sanitview.TViewStyle{Fg: g.red},
			Pre:             &sanitview.TViewStyle{Fg: g.fg4},
			FrameBackground: &sanitview.TViewStyle{Bg: bg},
			FrameHeader:     frameHeaders,
		},
	}
}
