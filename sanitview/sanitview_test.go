package sanitview

import (
	"reflect"
	"testing"
)

func TestStyles(t *testing.T) {
	t.Run("StringToStyle", func(t *testing.T) {
		s := StringToStyle("[a:b:c:d]")
		checkStylesEqual(t, &TViewStyle{
			Fg:    "a",
			Bg:    "b",
			Attrs: "c",
			Url:   "d",
		}, s)
		s = StringToStyle("[:::d]")
		checkStylesEqual(t, &TViewStyle{
			Fg:    "",
			Bg:    "",
			Attrs: "",
			Url:   "d",
		}, s)
	})
	t.Run("StyleToString", func(t *testing.T) {
		s := &TViewStyle{
			Fg:    "a",
			Bg:    "b",
			Attrs: "c",
			Url:   "d",
		}
		checkStringsEqual(t, "[a:b:c:d]", StyleToString(s))
		s = &TViewStyle{
			Fg:    "",
			Bg:    "",
			Attrs: "",
			Url:   "d",
		}
		checkStringsEqual(t, "[:::d]", StyleToString(s))
	})
	t.Run("MergeStyles", func(t *testing.T) {
		s1 := &TViewStyle{
			Fg:    "a",
			Bg:    "b",
			Attrs: "c",
			Url:   "d",
		}
		s2 := &TViewStyle{
			Fg:    "e",
			Bg:    "f",
			Attrs: "g",
			Url:   "h",
		}
		checkStylesEqual(t, s2, MergeTviewStyles(s1, s2))
		s2 = &TViewStyle{
			Fg:    "",
			Bg:    "f",
			Attrs: "g",
			Url:   "",
		}
		checkStylesEqual(t, &TViewStyle{
			Fg:    "a",
			Bg:    "f",
			Attrs: "g",
			Url:   "d",
		}, MergeTviewStyles(s1, s2))
	})

	// TODO AsTCellStyle
}

func checkStylesEqual(t *testing.T, e *TViewStyle, a *TViewStyle) {
	if !reflect.DeepEqual(a, e) {
		t.Errorf("Styles not equal\n  expected:%#v\n  actual:%#v\n", e, a)
	}
}

func checkStringsEqual(t *testing.T, e string, a string) {
	if a != e {
		t.Errorf("Strings not equal\n  expected:\"%s\"\n  actual:\"%s\"\n", e, a)
	}
}
