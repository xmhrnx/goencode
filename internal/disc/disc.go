// Package disc reads Blu-ray and DVD folders (a disc copied to disk, or
// the disc itself) and finds the main title, so a disc is added as one
// file instead of the hundreds of clips it's made of.
package disc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Kind is the type of disc.
type Kind int

const (
	BluRay Kind = iota + 1
	DVD
)

// String returns the kind for display.
func (k Kind) String() string {
	if k == DVD {
		return "DVD"
	}
	return "Blu-ray"
}

// Title is the main title of a disc: the files it plays, in order.
type Title struct {
	Kind     Kind
	Root     string // the disc's folder, which holds BDMV or VIDEO_TS
	Name     string
	Files    []string
	Size     int64
	Duration time.Duration // from the disc's playlists; 0 when unknown
	Source   string        // "00800.mpls", "VTS_01"

	// Languages are the audio and subtitle languages by stream PID, from
	// the playlist: Blu-ray clips seldom carry them.
	Languages map[int]string
}

// Root returns the disc folder that path is or is inside: path itself
// when it holds BDMV or VIDEO_TS, its parent when it's one of those, and
// "" when it isn't a disc.
func Root(path string) string {
	switch strings.ToUpper(filepath.Base(path)) {
	case "BDMV", "VIDEO_TS":
		if isDisc(filepath.Dir(path)) {
			return filepath.Dir(path)
		}
	}
	if isDisc(path) {
		return path
	}
	return ""
}

func isDisc(dir string) bool {
	if bd := child(dir, "BDMV"); bd != "" && child(bd, "PLAYLIST") != "" && child(bd, "STREAM") != "" {
		return true
	}
	if ts := child(dir, "VIDEO_TS"); ts != "" {
		ents, _ := os.ReadDir(ts)
		for _, e := range ents {
			if vobPattern.MatchString(strings.ToUpper(e.Name())) {
				return true
			}
		}
	}
	return false
}

// Open finds the main title of the disc at root (see Root).
func Open(root string) (*Title, error) {
	var t *Title
	var err error
	if bd := child(root, "BDMV"); bd != "" {
		t, err = openBluRay(bd)
	} else if ts := child(root, "VIDEO_TS"); ts != "" {
		t, err = openDVD(ts)
	} else {
		return nil, fmt.Errorf("%s isn't a Blu-ray or DVD folder", root)
	}
	if err != nil {
		return nil, err
	}
	t.Root = root
	if t.Name == "" {
		t.Name = filepath.Base(root)
		if t.Name == "" || t.Name == "." || t.Name == string(filepath.Separator) || strings.HasSuffix(t.Name, ":\\") {
			t.Name = t.Kind.String()
		}
	}
	for _, f := range t.Files {
		if fi, err := os.Stat(f); err == nil {
			t.Size += fi.Size()
		}
	}
	return t, nil
}

// Input is how ffmpeg reads the title: the path to pass to -i and the
// options that go before it. Several Blu-ray clips are joined by a
// concat list written to tmpDir, which keeps their timestamps in order;
// a DVD title's VOBs are one stream cut into 1 GB parts, joined as is.
// Both get a longer analysis, since subtitles start late in a film.
func (t *Title) Input(tmpDir string) (string, []string, error) {
	args := []string{"-probesize", "100M", "-analyzeduration", "100M"}
	switch {
	case len(t.Files) == 1:
		return t.Files[0], args, nil
	case t.Kind == DVD:
		return "concat:" + strings.Join(t.Files, "|"), args, nil
	}
	var b strings.Builder
	b.WriteString("ffconcat version 1.0\n")
	for _, f := range t.Files {
		abs, _ := filepath.Abs(f)
		fmt.Fprintf(&b, "file '%s'\n", strings.ReplaceAll(filepath.ToSlash(abs), "'", `'\''`))
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", nil, err
	}
	list := filepath.Join(tmpDir, "disc-"+strconv.FormatUint(hash(t.Root+t.Source), 36)+".ffconcat")
	if err := os.WriteFile(list, []byte(b.String()), 0o644); err != nil {
		return "", nil, err
	}
	return list, append([]string{"-f", "concat", "-safe", "0"}, args...), nil
}

func hash(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint64(s[i])) * 1099511628211
	}
	return h
}

// child finds a file or folder in dir by name, ignoring case: discs
// mounted on Linux can show names in lower case.
func child(dir, name string) string {
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

// Blu-ray

// openBluRay picks the playlist with the most distinct video whose clips
// are all present: the film, rather than trailers, menus and extras. A
// clip played over and over (a menu's background, looped for hours)
// counts once. Discs that hide the film among decoy playlists of the same
// length play the first of them.
func openBluRay(bdmv string) (*Title, error) {
	plDir, streamDir := child(bdmv, "PLAYLIST"), child(bdmv, "STREAM")
	ents, err := os.ReadDir(plDir)
	if err != nil {
		return nil, err
	}
	sort.Slice(ents, func(i, j int) bool { return ents[i].Name() < ents[j].Name() })
	var best *Title
	var bestDistinct time.Duration
	for _, e := range ents {
		if !strings.EqualFold(filepath.Ext(e.Name()), ".mpls") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(plDir, e.Name()))
		if err != nil {
			continue
		}
		pl, err := ParseMPLS(data)
		if err != nil || len(pl.Clips) == 0 {
			continue
		}
		files := make([]string, 0, len(pl.Clips))
		for _, c := range pl.Clips {
			f := child(streamDir, c+".m2ts")
			if f == "" {
				f = child(streamDir, c+".mts")
			}
			if f == "" {
				files = nil
				break
			}
			files = append(files, f)
		}
		if files == nil {
			continue
		}
		if best == nil || pl.Distinct > bestDistinct {
			best = &Title{Kind: BluRay, Files: files, Duration: pl.Duration, Source: e.Name(), Languages: pl.Languages}
			bestDistinct = pl.Distinct
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no playable title on this Blu-ray")
	}
	best.Name = bluRayName(bdmv)
	return best, nil
}

// Playlist is what a Blu-ray playlist plays.
type Playlist struct {
	Clips     []string // clip names, "00010", in order
	Duration  time.Duration
	Distinct  time.Duration  // the running time with each clip counted once
	Languages map[int]string // by PID, from the first play item
}

// ParseMPLS reads a Blu-ray playlist: the clips it plays, in order, how
// long it runs and the languages of its audio and subtitle streams.
func ParseMPLS(b []byte) (Playlist, error) {
	var pl Playlist
	if len(b) < 20 || string(b[:4]) != "MPLS" {
		return pl, fmt.Errorf("not a Blu-ray playlist")
	}
	start := int(binary.BigEndian.Uint32(b[8:12]))
	if start+10 > len(b) {
		return pl, fmt.Errorf("playlist is cut short")
	}
	n := int(binary.BigEndian.Uint16(b[start+6:]))
	p := start + 10
	var ticks, distinct uint64 // 45 kHz
	seen := map[string]bool{}
	for i := 0; i < n; i++ {
		if p+22 > len(b) {
			return pl, fmt.Errorf("playlist is cut short")
		}
		size := int(binary.BigEndian.Uint16(b[p:]))
		end := p + 2 + size
		clip := string(b[p+2 : p+7])
		pl.Clips = append(pl.Clips, clip)
		in := binary.BigEndian.Uint32(b[p+14:])
		out := binary.BigEndian.Uint32(b[p+18:])
		if out > in {
			ticks += uint64(out - in)
			if !seen[clip] {
				distinct += uint64(out - in)
			}
		}
		seen[clip] = true
		if i == 0 && end <= len(b) {
			pl.Languages = stnLanguages(b[:end], p)
		}
		p = end
	}
	pl.Duration = time.Duration(ticks) * time.Second / 45000
	pl.Distinct = time.Duration(distinct) * time.Second / 45000
	return pl, nil
}

// stnLanguages reads the stream table of the play item at p: each
// stream's PID and, for audio and subtitles, its language code.
func stnLanguages(b []byte, p int) map[int]string {
	q := p + 34 // past the play item's fixed fields
	if q > len(b) {
		return nil
	}
	if b[p+12]&0x10 != 0 { // multi-angle: skip the other angles' clips
		if q+2 > len(b) {
			return nil
		}
		q += 2 + (int(b[q])-1)*10
	}
	if q+16 > len(b) {
		return nil
	}
	count := 0
	for _, c := range b[q+4 : q+11] {
		count += int(c)
	}
	q += 16
	langs := map[int]string{}
	for i := 0; i < count && q < len(b); i++ {
		// stream_entry: its type, then the PID at a type-dependent place.
		entry := q + 1
		if entry+int(b[q]) > len(b) || b[q] < 3 {
			break
		}
		pid := -1
		switch b[entry] {
		case 1:
			pid = int(binary.BigEndian.Uint16(b[entry+1:]))
		case 2:
			pid = int(binary.BigEndian.Uint16(b[entry+3:]))
		case 3, 4:
			pid = int(binary.BigEndian.Uint16(b[entry+2:]))
		}
		q = entry + int(b[q])
		// stream_attributes: the coding type, then the language for
		// audio (after a format byte), subtitles and menus.
		if q >= len(b) {
			break
		}
		attr, alen := q+1, int(b[q])
		if attr+alen > len(b) || alen < 1 {
			break
		}
		lang := ""
		switch t := b[attr]; {
		case (t >= 0x80 && t <= 0x86) || t == 0xA1 || t == 0xA2:
			if alen >= 5 {
				lang = string(b[attr+2 : attr+5])
			}
		case t == 0x90 || t == 0x91:
			if alen >= 4 {
				lang = string(b[attr+1 : attr+4])
			}
		case t == 0x92:
			if alen >= 5 {
				lang = string(b[attr+2 : attr+5])
			}
		}
		if pid >= 0 && lang != "" && strings.Trim(lang, "\x00 ") != "" {
			langs[pid] = strings.ToLower(lang)
		}
		q = attr + alen
	}
	return langs
}

var bdName = regexp.MustCompile(`<di:name>([^<]+)</di:name>`)

// bluRayName is the title in the disc's metadata, when it has some.
func bluRayName(bdmv string) string {
	meta := child(bdmv, "META")
	if meta == "" {
		return ""
	}
	dl := child(meta, "DL")
	if dl == "" {
		return ""
	}
	ents, _ := os.ReadDir(dl)
	for _, e := range ents {
		if !strings.HasPrefix(strings.ToLower(e.Name()), "bdmt_") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dl, e.Name()))
		if err != nil {
			continue
		}
		if m := bdName.FindSubmatch(data); m != nil {
			return strings.TrimSpace(string(m[1]))
		}
	}
	return ""
}

// DVD

var vobPattern = regexp.MustCompile(`^VTS_(\d\d)_([1-9])\.VOB$`)

// openDVD picks the title set with the most video: the film. Its VOBs
// after VTS_NN_0 (the menus) are the film in order.
func openDVD(videoTS string) (*Title, error) {
	ents, err := os.ReadDir(videoTS)
	if err != nil {
		return nil, err
	}
	type set struct {
		files []string
		size  int64
	}
	sets := map[string]*set{}
	for _, e := range ents {
		m := vobPattern.FindStringSubmatch(strings.ToUpper(e.Name()))
		if m == nil {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		s := sets[m[1]]
		if s == nil {
			s = &set{}
			sets[m[1]] = s
		}
		s.files = append(s.files, filepath.Join(videoTS, e.Name()))
		s.size += fi.Size()
	}
	best := ""
	for id, s := range sets {
		if best == "" || s.size > sets[best].size || (s.size == sets[best].size && id < best) {
			best = id
		}
	}
	if best == "" {
		return nil, fmt.Errorf("no playable title on this DVD")
	}
	files := sets[best].files
	sort.Slice(files, func(i, j int) bool { return strings.ToUpper(files[i]) < strings.ToUpper(files[j]) })
	t := &Title{Kind: DVD, Files: files, Source: "VTS_" + best}
	if ifo := child(videoTS, "VTS_"+best+"_0.IFO"); ifo != "" {
		if data, err := os.ReadFile(ifo); err == nil {
			t.Duration, _ = ParseIFO(data)
		}
	}
	return t, nil
}

// ParseIFO reads a DVD title set's IFO and returns the running time of
// its longest program chain: the film's length.
func ParseIFO(b []byte) (time.Duration, error) {
	if len(b) < 0xD0 || string(b[:12]) != "DVDVIDEO-VTS" {
		return 0, fmt.Errorf("not a DVD title set")
	}
	base := int(binary.BigEndian.Uint32(b[0xCC:])) * 2048
	if base == 0 || base+8 > len(b) {
		return 0, fmt.Errorf("title set has no program chains")
	}
	n := int(binary.BigEndian.Uint16(b[base:]))
	var longest time.Duration
	for i := 0; i < n; i++ {
		srp := base + 8 + i*8
		if srp+8 > len(b) {
			break
		}
		pgc := base + int(binary.BigEndian.Uint32(b[srp+4:]))
		if pgc+8 > len(b) {
			continue
		}
		if d := bcdTime(b[pgc+4 : pgc+8]); d > longest {
			longest = d
		}
	}
	return longest, nil
}

// bcdTime decodes a DVD playback time: hours, minutes, seconds and
// frames in BCD, with the frame rate in the top bits of the last byte.
func bcdTime(t []byte) time.Duration {
	bcd := func(v byte) int { return int(v>>4)*10 + int(v&0x0f) }
	d := time.Duration(bcd(t[0]))*time.Hour + time.Duration(bcd(t[1]))*time.Minute + time.Duration(bcd(t[2]))*time.Second
	fps := 25.0
	if t[3]>>6 == 3 {
		fps = 29.97
	}
	return d + time.Duration(float64(bcd(t[3]&0x3f))/fps*float64(time.Second))
}
