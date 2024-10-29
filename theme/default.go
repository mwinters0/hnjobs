package theme

import "hnjobs/sanitview"

// only used to populate default-theme.json
func getDefaultTheme() Theme {
	white := "#f9f5d7"
	black := "#1d2021"
	hnOrange := "#cc5200"
	normal := &sanitview.TViewStyle{Fg: white, Bg: black}
	companyNameFg := "deepskyblue"
	return Theme{
		Version: 1,
		UI: UIColors{
			HeaderStatsDate:   &sanitview.TViewStyle{Fg: white, Bg: hnOrange},
			HeaderStatsNormal: &sanitview.TViewStyle{Fg: black, Bg: "orange", Attrs: "b"},
			HeaderStatsHidden: &sanitview.TViewStyle{Fg: "#BBBBBB", Bg: "#333333"},
			FocusBorder:       &sanitview.TViewStyle{Fg: white},
			ModalTitle:        &sanitview.TViewStyle{Fg: black, Bg: "orange"},
			ModalNormal:       normal,
			ModalHighlight:    &sanitview.TViewStyle{Fg: "orange", Bg: black},
		},
		CompanyList: CompanyList{
			Colors: CompanyListColors{
				CompanyName:             &sanitview.TViewStyle{Fg: companyNameFg, Bg: black},
				CompanyNameUnread:       &sanitview.TViewStyle{Attrs: "b"},
				CompanyNamePriority:     &sanitview.TViewStyle{Fg: black, Bg: "orange"},
				CompanyNameApplied:      &sanitview.TViewStyle{},
				CompanyNameUninterested: &sanitview.TViewStyle{Fg: "grey"},
				StatusChar:              &sanitview.TViewStyle{Fg: white, Bg: black},
				StatusCharUnread:        &sanitview.TViewStyle{Fg: "purple", Attrs: "b"},
				StatusCharPriority:      &sanitview.TViewStyle{Fg: black, Bg: "orange"},
				StatusCharApplied:       &sanitview.TViewStyle{Fg: "green"},
				StatusCharUninterested:  &sanitview.TViewStyle{Fg: "grey"},
				Score:                   &sanitview.TViewStyle{Fg: white, Bg: black},
				ScoreUnread:             &sanitview.TViewStyle{Attrs: "b"},
				ScorePriority:           &sanitview.TViewStyle{Fg: "black", Bg: "orange"},
				ScoreApplied:            &sanitview.TViewStyle{},
				ScoreUninterested:       &sanitview.TViewStyle{Fg: "grey"},
				SelectedItemBackground:  &sanitview.TViewStyle{Bg: "#444444"},
				FrameBackground:         &sanitview.TViewStyle{Bg: black},
				FrameHeader:             &sanitview.TViewStyle{Fg: white, Bg: hnOrange},
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
			URL:             &sanitview.TViewStyle{Fg: "#c69749", Attrs: "u"},
			Email:           &sanitview.TViewStyle{Fg: "pink"},
			PositiveHit:     &sanitview.TViewStyle{Fg: "#a1f6ae"},
			NegativeHit:     &sanitview.TViewStyle{Fg: "red"},
			Pre:             &sanitview.TViewStyle{Fg: "#888888"},
			FrameBackground: &sanitview.TViewStyle{Bg: black},
			FrameHeader:     &sanitview.TViewStyle{Fg: white, Bg: hnOrange},
		},
	}
}
