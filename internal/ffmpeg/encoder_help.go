package ffmpeg

import (
	"context"
	"regexp"
	"strings"
)

// EncoderHelp is what `ffmpeg -h encoder=NAME` says about an encoder.
type EncoderHelp struct {
	Name          string
	Description   string
	Hardware      bool
	PixelFormats  []string
	SampleRates   []string
	SampleFormats []string
	Layouts       []string
	Options       []AVOption
}

// AVOption is one of the encoder's private options.
type AVOption struct {
	Name        string
	Type        string // "int", "float", "boolean", "string", "flags", ...
	Description string
	Default     string
	Min, Max    string
	Constants   []AVConstant
}

// AVConstant is a named value an option accepts.
type AVConstant struct {
	Name        string
	Value       string
	Description string
}

// EncoderHelp returns the parsed help of an encoder, running ffmpeg the
// first time and caching the result. Concurrent callers share one run.
func (c *Capabilities) EncoderHelp(ctx context.Context, enc string) *EncoderHelp {
	c.mu.Lock()
	if h, ok := c.encHelp[enc]; ok {
		c.mu.Unlock()
		return h
	}
	if w, ok := c.helpWait[enc]; ok {
		c.mu.Unlock()
		<-w
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.encHelp[enc]
	}
	w := make(chan struct{})
	c.helpWait[enc] = w
	c.mu.Unlock()

	h := &EncoderHelp{Name: enc}
	if out, err := command(ctx, c.Binaries.FFmpeg, "-hide_banner", "-h", "encoder="+enc).Output(); err == nil {
		h = parseEncoderHelp(enc, string(out))
	}
	c.mu.Lock()
	c.encHelp[enc] = h
	delete(c.helpWait, enc)
	c.mu.Unlock()
	close(w)
	return h
}

// CachedEncoderHelp returns the help if it was loaded already, without
// running ffmpeg: for the UI goroutine, which mustn't block.
func (c *Capabilities) CachedEncoderHelp(enc string) *EncoderHelp {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.encHelp[enc]
}

var (
	optionLine   = regexp.MustCompile(`^  -(\S+)\s+<(\w+)>\s+[A-Z.]{8,}\s?(.*)$`)
	constantLine = regexp.MustCompile(`^\s{4,}(\S+)\s+(\S+)\s+[A-Z.]{8,}\s?(.*)$`)
	rangeRe      = regexp.MustCompile(`\(from (\S+) to (\S+)\)`)
	defaultRe    = regexp.MustCompile(`\(default (.+?)\)\s*$`)
)

func parseEncoderHelp(enc, out string) *EncoderHelp {
	h := &EncoderHelp{Name: enc}
	var cur *AVOption
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r ")
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Encoder "):
			if i := strings.Index(line, "["); i >= 0 {
				h.Description = strings.TrimSuffix(strings.TrimSuffix(line[i+1:], ":"), "]")
			}
		case strings.HasPrefix(trim, "General capabilities:"):
			h.Hardware = strings.Contains(trim, "hardware")
		case strings.HasPrefix(trim, "Supported pixel formats:"):
			h.PixelFormats = strings.Fields(strings.TrimPrefix(trim, "Supported pixel formats:"))
		case strings.HasPrefix(trim, "Supported sample rates:"):
			h.SampleRates = strings.Fields(strings.TrimPrefix(trim, "Supported sample rates:"))
		case strings.HasPrefix(trim, "Supported sample formats:"):
			h.SampleFormats = strings.Fields(strings.TrimPrefix(trim, "Supported sample formats:"))
		case strings.HasPrefix(trim, "Supported channel layouts:"):
			h.Layouts = strings.Fields(strings.TrimPrefix(trim, "Supported channel layouts:"))
		default:
			if m := optionLine.FindStringSubmatch(line); m != nil {
				o := AVOption{Name: m[1], Type: m[2]}
				desc := m[3]
				if d := defaultRe.FindStringSubmatch(desc); d != nil {
					o.Default = strings.Trim(d[1], `"`)
					desc = strings.TrimSpace(desc[:len(desc)-len(d[0])])
				}
				if r := rangeRe.FindStringSubmatch(desc); r != nil {
					o.Min, o.Max = r[1], r[2]
					desc = strings.TrimSpace(strings.Replace(desc, r[0], "", 1))
				}
				o.Description = desc
				h.Options = append(h.Options, o)
				cur = &h.Options[len(h.Options)-1]
				continue
			}
			if cur != nil {
				if m := constantLine.FindStringSubmatch(line); m != nil {
					cur.Constants = append(cur.Constants, AVConstant{Name: m[1], Value: m[2], Description: strings.TrimSpace(m[3])})
				}
			}
		}
	}
	return h
}
