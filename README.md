[![Go Report](https://goreportcard.com/badge/github.com/mwinters0/hnjobs)](https://goreportcard.com/report/github.com/mwinters0/hnjobs)

# hnjobs
A console tool to find your best match on Who's Hiring.  Exports to JSON for automated ennui. It:
1. Finds the latest Who's Hiring post and fetches / caches all job comments locally in sqlite.
2. Scores job postings according to your criteria.
3. Provides a TUI to help you review the listings and track which ones are interesting / applied to / ruled out.

## Installation
```shell
go install github.com/mwinters0/hnjobs
```

Or grab a binary from the [releases](https://github.com/mwinters0/hnjobs/releases)

## Usage
On first run, a config file is created at `UserConfigDir/hnjobs/config.json`.  (On linux this is
`~/.config/hnjobs/config.json`.) 👉 **You should edit the config file**👈 before running any other commands, as this is
where your scoring rules are stored.  Some samples rules are provided in the generated file.  Each rule is a
[golang regex](https://pkg.go.dev/regexp/syntax) which must be JSON escaped (`\b` -> `\\b`).

After you've set up your rules, run `hnjobs` again and it will auto-fetch the most-recent job story, score the jobs by
your criteria, and show the TUI.

### TUI bindings
- Basics
  - `ESC` - close dialogs
  - `TAB` - switch focus (so you can scroll a long job listing if needed)
  - `jk` and up/down arrows - scroll
    - `g`, `G`, `Ctrl-d`, `Ctrl-u` - scroll harder 
  - `f` - fetch latest (only fetches new / TTL expired jobs, with default TTL of 1 day)
  - `F` - force fetch all jobs (ignore TTL)
  - `q` - quit
- Job Filtering
  - `r` - mark read / unread
  - `x` - mark job uninterested (hidden) or interested (default)
  - `p` - mark priority / not priority
  - `a` - mark applied to / not applied to
- Display
  - `X` - toggle hiding of jobs marked uninterested
  - `T` - toggle hiding of jobs below your score threshold (set in the config file)
  - `m` - select month (if multiple in your DB) / delete old months
  - `s` - reload config file and re-score the jobs (useful if you changed your rules)

### Commands
```shell
hnjobs # Works offline.
hnjobs fetch # Just fetch, no TUI. Run this before hopping on a plane.
hnjobs fetch -x # Fetch and set exit code according to results. 0 = new jobs available. #bashlife
hnjobs rescore # Re-score the cached jobs by your criteria. Only needed if you've changed your rules.
hnjobs dump # Dump the current month's data to JSON on stdout.
```

## Scoring rules FAQ
- `text_missing` rules match if the regex fails.  Use this to influence the score if a word is missing from a listing.
- `why` and `why_not` tags are optional.  I like to analyze my past decisions whenever I watch my credit score drop. 🤷  These will 
become visible in the TUI eventually.

## Styling
If you hate orange, you can edit your config file to use one of the built-in themes: `material` or
`gruvbox[dark|light]<hard|soft>` (example: `gruvboxdarkhard`).

If you want to take a crack at making your own, have a look at the default theme which is generated on first-run at
`UserConfigDir/hnjobs/theme-default.json` (on linux: `~/.config/hnjobs/theme-default.json`).  You can either edit this
or copy it to `theme-foo.json` and set your config's theme to `foo`.

## Misc
The database is stored at `UserDataDir/hnjobs/hnjobs.sqlite` (on linux: `~/.local/share/hnjobs/hnjobs.sqlite`).  
