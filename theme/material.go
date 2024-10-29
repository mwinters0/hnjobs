package theme

import "github.com/mwinters0/hnjobs/sanitview"

func getMaterialTheme() Theme {
	// converted from https://github.com/MartinSeeler/iterm2-material-design/blob/master/material-design-colors.itermcolors
	var material = struct {
		ansi0      string // bluegrey
		ansi1      string // red
		ansi2      string // brightgreen
		ansi3      string // brightyellow
		ansi4      string // brightblue
		ansi5      string // magenta
		ansi6      string // brightmint
		ansi7      string // almostwhite
		ansi8      string // midgrey
		ansi9      string // brightsalmon
		ansi10     string // mint
		ansi11     string // yellow
		ansi12     string // lightblue
		ansi13     string // salmon
		ansi14     string // brightmint
		ansi15     string // almost white
		background string // almost black
		bold       string // paper
		foreground string // brightpaper
		selection  string // union
	}{
		ansi0:      "#435a66",
		ansi1:      "#fb3841",
		ansi2:      "#5cf09e",
		ansi3:      "#fed032",
		ansi4:      "#36b6fe",
		ansi5:      "#fb216e",
		ansi6:      "#58ffd1",
		ansi7:      "#fffefe",
		ansi8:      "#a0b0b8",
		ansi9:      "#fc736d",
		ansi10:     "#acf6be",
		ansi11:     "#fee16c",
		ansi12:     "#6fcefe",
		ansi13:     "#fc669a",
		ansi14:     "#99ffe5",
		ansi15:     "#fffefe",
		background: "#1c252a",
		bold:       "#e9e9e9",
		foreground: "#e7eaed",
		selection:  "#4e6978",
	}

	white := material.foreground
	black := material.background
	brandStrong := material.ansi6
	companyNameFg := material.ansi4

	normal := &sanitview.TViewStyle{Fg: white, Bg: black}
	dim := &sanitview.TViewStyle{Fg: material.ansi8}
	priority := &sanitview.TViewStyle{Fg: material.background, Bg: material.ansi14}
	frameHeader := &sanitview.TViewStyle{Fg: white, Bg: material.selection}

	return Theme{
		Version: 1,
		UI: UIColors{
			HeaderStatsDate:   &sanitview.TViewStyle{Fg: material.background, Bg: material.ansi10},
			HeaderStatsNormal: &sanitview.TViewStyle{Fg: material.background, Bg: material.ansi12},
			HeaderStatsHidden: &sanitview.TViewStyle{Fg: material.ansi8, Bg: material.ansi0},
			FocusBorder:       &sanitview.TViewStyle{Fg: brandStrong},
			ModalTitle:        &sanitview.TViewStyle{Fg: material.ansi11, Bg: material.selection},
			ModalNormal:       normal,
			ModalHighlight:    &sanitview.TViewStyle{Fg: brandStrong, Bg: black},
		},
		CompanyList: CompanyList{
			Colors: CompanyListColors{
				CompanyName:             &sanitview.TViewStyle{Fg: companyNameFg, Bg: black},
				CompanyNameUnread:       &sanitview.TViewStyle{Attrs: "b"},
				CompanyNamePriority:     priority,
				CompanyNameApplied:      &sanitview.TViewStyle{},
				CompanyNameUninterested: dim,
				StatusChar:              normal,
				StatusCharUnread:        &sanitview.TViewStyle{Fg: material.ansi13, Attrs: "b"},
				StatusCharPriority:      priority,
				StatusCharApplied:       &sanitview.TViewStyle{Fg: material.ansi10},
				StatusCharUninterested:  dim,
				Score:                   normal,
				ScoreUnread:             &sanitview.TViewStyle{Attrs: "b"},
				ScorePriority:           priority,
				ScoreApplied:            &sanitview.TViewStyle{},
				ScoreUninterested:       dim,
				SelectedItemBackground:  &sanitview.TViewStyle{Bg: material.selection},
				FrameBackground:         &sanitview.TViewStyle{Bg: black},
				FrameHeader:             frameHeader,
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
			CompanyName:     &sanitview.TViewStyle{Fg: companyNameFg},
			URL:             &sanitview.TViewStyle{Fg: material.ansi10, Attrs: "u"},
			Email:           &sanitview.TViewStyle{Fg: material.ansi14},
			PositiveHit:     &sanitview.TViewStyle{Fg: material.ansi6},
			NegativeHit:     &sanitview.TViewStyle{Fg: material.ansi9},
			Pre:             &sanitview.TViewStyle{Fg: material.ansi8},
			FrameBackground: &sanitview.TViewStyle{Bg: black},
			FrameHeader:     frameHeader,
		},
	}

}
