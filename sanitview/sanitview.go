// Package sanitview provides foundational capabilities for managing tview styles.  It enables them to be built
// programmatically, and creates functionality similar to cascading stylesheets where one style can override a subset
// of another.
package sanitview

import (
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"strings"
)

type TViewStringStyle = string
type TViewStyle struct {
	Fg    string
	Bg    string
	Attrs string
	Url   string
}

func (tvs *TViewStyle) AsTag() string {
	return StyleToString(tvs)
}

func (tvs *TViewStyle) AsTCellStyle() tcell.Style {
	s := tcell.Style{}
	if tvs.Fg != "" {
		s = s.Foreground(tcell.GetColor(tvs.Fg))
	}
	if tvs.Bg != "" {
		s = s.Background(tcell.GetColor(tvs.Bg))
	}
	if tvs.Attrs != "" {
		a := strings.ToLower(tvs.Attrs)
		var m tcell.AttrMask
		if strings.Contains(a, "l") {
			m ^= tcell.AttrBlink
		}
		if strings.Contains(a, "b") {
			m ^= tcell.AttrBold
		}
		if strings.Contains(a, "i") {
			m ^= tcell.AttrItalic
		}
		if strings.Contains(a, "d") {
			m ^= tcell.AttrDim
		}
		if strings.Contains(a, "r") {
			m ^= tcell.AttrReverse
		}
		if strings.Contains(a, "u") {
			m ^= tcell.AttrUnderline
		}
		if strings.Contains(a, "s") {
			m ^= tcell.AttrStrikeThrough
		}
		s = s.Attributes(m)
	}
	return s
}

func (tvs *TViewStyle) Clone() *TViewStyle {
	return &TViewStyle{
		Fg:    tvs.Fg,
		Bg:    tvs.Bg,
		Attrs: tvs.Attrs,
		Url:   tvs.Url,
	}
}

func (s *TViewStyle) IsEmpty() bool {
	return s.Fg == "" && s.Bg == "" && s.Attrs == "" && s.Url == ""
}

func MergeTviewStyles(styles ...*TViewStyle) *TViewStyle {
	if len(styles) < 2 {
		panic("styles must have at least two elements")
	}
	style := TViewStyle{}
	for i, newStyle := range styles {
		if i == 0 {
			style = *newStyle
			continue
		}
		if newStyle.Fg != "" {
			style.Fg = newStyle.Fg
		}
		if newStyle.Bg != "" {
			style.Bg = newStyle.Bg
		}
		if newStyle.Attrs != "" {
			style.Attrs = newStyle.Attrs
		}
		if newStyle.Url != "" {
			style.Url = newStyle.Url
		}
	}
	return &style
}

func StringToStyle(s TViewStringStyle) *TViewStyle {
	style := &TViewStyle{}
	if s[0:1] != "[" {
		panic(fmt.Errorf("not a tview string style: '%s'", s))
	}
	if s[len(s)-1:] != "]" {
		panic(fmt.Errorf("not a tview string style: '%s'", s))
	}
	if !strings.ContainsAny(s, ":") {
		// Fg only
		style.Fg = s[1 : len(s)-1]
		return style
	}
	i := strings.IndexRune(s, ':')
	style.Fg = s[1:i]
	s = s[i+1:]

	i = strings.IndexRune(s, ':')
	if i == -1 {
		style.Bg = s[:len(s)-1]
		return style
	}
	style.Bg = s[:i]
	s = s[i+1:]

	i = strings.IndexRune(s, ':')
	if i == -1 {
		style.Attrs = s[:len(s)-1]
		return style
	}
	style.Attrs = s[:i]
	s = s[i+1:]

	if len(s) > 1 {
		//nothing left but Url
		style.Url = s[0 : len(s)-1]
		return style
	}

	panic(errors.New("couldn't find end of tview string"))
}

func StyleToString(style *TViewStyle) TViewStringStyle {
	if style.IsEmpty() {
		return ""
	}
	s := "[" + strings.Join([]string{style.Fg, style.Bg, style.Attrs, style.Url}, ":") + "]"
	return s
}
