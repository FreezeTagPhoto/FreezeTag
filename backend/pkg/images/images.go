package images

import (
	"fmt"
	"time"

	"github.com/gobwas/glob"
)

// standardized image information
type Info struct {
	// image data, obviously will always appear
	image struct {
		pixelsRGBA []byte
		width      int
		height     int
	}
	// image metadata, only sometimes appears in EXIF-format images
	// and even then only parts may appear
	meta *struct {
		// geolocation information
		geo *struct {
			location string
			time     time.Time
		}
		// manufacturer information
		cam *struct {
			manufacturer string
			model        string
		}
	}
}

type Parser interface {
	ParseImage(name string, data []byte) (Info, error)
}

type ParserFunc func(string, []byte) (Info, error)

func (p ParserFunc) ParseImage(name string, data []byte) (Info, error) {
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
func (pc *ParserCollection) RegisterParserFunc(match string, parser func(string, []byte) (Info, error)) error {
	return pc.RegisterParser(match, ParserFunc(parser))
}

func (pc ParserCollection) ParseImage(name string, data []byte) (Info, error) {
	for _, entry := range pc.parsers {
		if entry.matcher.Match(name) {
			return entry.parser.ParseImage(name, data)
		}
	}
	return Info{}, fmt.Errorf("no parser for file: %q", name)
}
