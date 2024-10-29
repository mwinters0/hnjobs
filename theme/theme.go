package theme

import (
	"encoding/json"
	"fmt"
	"hnjobs/sanitview"
	"os"
	"path/filepath"
	"strings"
)

var curTheme Theme

type Theme struct {
	Version     int
	UI          UIColors
	CompanyList CompanyList
	JobBody     JobBodyColors
}

type CompanyList struct {
	Chars  CompanyListChars
	Colors CompanyListColors
}

type CompanyListChars struct {
	Unread       string
	Read         string
	Applied      string
	Uninterested string
	Priority     string
}

type CompanyListColors struct {
	CompanyName             *sanitview.TViewStyle
	CompanyNameUnread       *sanitview.TViewStyle
	CompanyNamePriority     *sanitview.TViewStyle
	CompanyNameApplied      *sanitview.TViewStyle
	CompanyNameUninterested *sanitview.TViewStyle
	StatusChar              *sanitview.TViewStyle
	StatusCharUnread        *sanitview.TViewStyle
	StatusCharPriority      *sanitview.TViewStyle
	StatusCharApplied       *sanitview.TViewStyle
	StatusCharUninterested  *sanitview.TViewStyle
	Score                   *sanitview.TViewStyle
	ScoreUnread             *sanitview.TViewStyle
	ScorePriority           *sanitview.TViewStyle
	ScoreApplied            *sanitview.TViewStyle
	ScoreUninterested       *sanitview.TViewStyle
	SelectedItemBackground  *sanitview.TViewStyle
	FrameBackground         *sanitview.TViewStyle
	FrameHeader             *sanitview.TViewStyle
}

type JobBodyColors struct {
	Normal          *sanitview.TViewStyle
	CompanyName     *sanitview.TViewStyle
	URL             *sanitview.TViewStyle
	Email           *sanitview.TViewStyle
	PositiveHit     *sanitview.TViewStyle
	NegativeHit     *sanitview.TViewStyle
	Pre             *sanitview.TViewStyle
	FrameBackground *sanitview.TViewStyle
	FrameHeader     *sanitview.TViewStyle
}

type UIColors struct {
	HeaderStatsDate   *sanitview.TViewStyle
	HeaderStatsNormal *sanitview.TViewStyle
	HeaderStatsHidden *sanitview.TViewStyle
	FocusBorder       *sanitview.TViewStyle
	ModalTitle        *sanitview.TViewStyle
	ModalNormal       *sanitview.TViewStyle
	ModalHighlight    *sanitview.TViewStyle
}

// GetTheme returns the currently-loaded theme
func GetTheme() Theme {
	return curTheme
}

func LoadTheme(name string, path string) error {
	if name == "" {
		name = "default"
	}
	name = strings.Replace(name, "/", "", -1) //bunny foo foo
	switch strings.ToLower(name) {
	case "material":
		curTheme = getMaterialTheme()
	case "gruvbox", "gruvboxdark":
		curTheme = getGruvboxTheme(GruvboxModeDark, GruvboxIntensityNeutral)
	case "gruvboxdarkhard":
		curTheme = getGruvboxTheme(GruvboxModeDark, GruvboxIntensityHard)
	case "gruvboxdarksoft":
		curTheme = getGruvboxTheme(GruvboxModeDark, GruvboxIntensitySoft)
	case "gruvboxlight":
		curTheme = getGruvboxTheme(GruvboxModeLight, GruvboxIntensityNeutral)
	case "gruvboxlighthard":
		curTheme = getGruvboxTheme(GruvboxModeLight, GruvboxIntensityHard)
	case "gruvboxlightsoft":
		curTheme = getGruvboxTheme(GruvboxModeLight, GruvboxIntensitySoft)
	default:
		fullPath := filepath.Join(path, "theme-"+name+".json")
		err := loadThemeFile(fullPath)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func loadThemeFile(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file \"%s\": %v", filename, err)
	}
	err = loadThemeJSON(contents)
	if err != nil {
		return fmt.Errorf("error loading file \"%s\": %v", filename, err)
	}
	return nil
}

func loadThemeJSON(j []byte) error {
	curTheme = Theme{}
	err := json.Unmarshal(j, &curTheme)
	if err != nil {
		return err
	}
	// TODO validation ...
	return nil
}

func DefaultThemeFileContents() []byte {
	j, err := json.MarshalIndent(getDefaultTheme(), "", "  ")
	if err != nil {
		panic(err)
	}
	return j
}
