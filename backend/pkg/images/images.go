package images

import (
	"fmt"
	"time"

	"github.com/gobwas/glob"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func init() {
	imagick.Initialize()
}

// standardized image information
type Data struct {
	// image data, obviously will always appear
	PixelsRGBA []byte
	Width      int
	Height     int
	// image metadata, only some parts of this may appear
	// otherwise they'll be nil
	DateCreated *time.Time
	Geo         *struct {
		Lat float64
		Lon float64
		Alt float64
	}
	Cam *struct {
		Manufacturer string
		Model        string
	}
}

type Parser interface {
	ParseImage(name string, data []byte) (Data, error)
}

type ParserFunc func(string, []byte) (Data, error)

func (p ParserFunc) ParseImage(name string, data []byte) (Data, error) {
	return p(name, data)
}

type parserEntry struct {
	matcher glob.Glob
	parser  Parser
}

type ParserCollection struct {
	parsers []parserEntry
}

// Create a new empty image parser collection
func InitParserCollection() ParserCollection {
	return ParserCollection{}
}

// Register an image parser and activation glob with this parser collection
//
// Globs are defined using syntax you can find at https://github.com/gobwas/glob.
func (pc *ParserCollection) RegisterParser(match string, parser Parser) error {
	matcher, err := glob.Compile(match)
	if err != nil {
		return fmt.Errorf("failed to compile glob: %w", err)
	}
	pc.parsers = append(pc.parsers, parserEntry{
		matcher,
		parser,
	})
	return nil
}

// Register an image parser function and activation glob with this parser collection
//
// Globs are defined using syntax you can find at https://github.com/gobwas/glob.
func (pc *ParserCollection) RegisterParserFunc(match string, parser func(string, []byte) (Data, error)) error {
	return pc.RegisterParser(match, ParserFunc(parser))
}

func (pc ParserCollection) ParseImage(name string, data []byte) (Data, error) {
	for _, entry := range pc.parsers {
		if entry.matcher.Match(name) {
			return entry.parser.ParseImage(name, data)
		}
	}
	return Data{}, fmt.Errorf("no parser for file: %q", name)
}
