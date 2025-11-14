package images

import (
	"fmt"
	"freezetag/backend/pkg/images/imagedata"
	"log"

	"github.com/gobwas/glob"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func init() {
	log.Println("[INFO] initializing ImageMagick...")
	imagick.Initialize()
	log.Println("[INFO] ImageMagick finished initializing.")
}

type Parser interface {
	ParseImage(name string, data []byte) (imagedata.Data, error)
}

type ParserFunc func(string, []byte) (imagedata.Data, error)

func (p ParserFunc) ParseImage(name string, data []byte) (imagedata.Data, error) {
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
func (pc *ParserCollection) RegisterParserFunc(match string, parser func(string, []byte) (imagedata.Data, error)) error {
	return pc.RegisterParser(match, ParserFunc(parser))
}

func (pc ParserCollection) ParseImage(name string, data []byte) (imagedata.Data, error) {
	for _, entry := range pc.parsers {
		if entry.matcher.Match(name) {
			return entry.parser.ParseImage(name, data)
		}
	}
	return imagedata.Data{}, fmt.Errorf("no parser for file")
}
