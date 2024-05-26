package feed

import (
	"bytes"
	"errors"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
	xpp "github.com/mmcdole/goxpp"
	"golang.org/x/net/html/charset"
)

// FeedType represents one of the possible feed
// types that we can detect.
type FeedType int

const (
	// FeedTypeUnknown represents a feed that could not have its
	// type determiend.
	FeedTypeUnknown FeedType = iota
	// FeedTypeAtom repesents an Atom feed
	FeedTypeAtom
	// FeedTypeRSS represents an RSS feed
	FeedTypeRSS
	// FeedTypeJSON represents a JSON feed
	FeedTypeJSON
	// pentesterland json feed
	FeedTypePenLandJSON
)

// DetectFeedType attempts to determine the type of feed
// by looking for specific xml elements unique to the
// various feed types.
func DetectFeedType(feed io.Reader) FeedType {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(feed)

	var firstChar byte
loop:
	for {
		ch, err := buffer.ReadByte()
		if err != nil {
			return FeedTypeUnknown
		}
		// ignore leading whitespace & byte order marks
		switch ch {
		case ' ', '\r', '\n', '\t':
		case 0xFE, 0xFF, 0x00, 0xEF, 0xBB, 0xBF: // utf 8-16-32 bom
		default:
			firstChar = ch
			buffer.UnreadByte()
			break loop
		}
	}

	if firstChar == '<' {
		// Check if it's an XML based feed
		p := xpp.NewXMLPullParser(bytes.NewReader(buffer.Bytes()), false, NewReaderLabel)

		_, err := FindRoot(p)
		if err != nil {
			return FeedTypeUnknown
		}

		name := strings.ToLower(p.Name)
		switch name {
		case "rdf":
			return FeedTypeRSS
		case "rss":
			return FeedTypeRSS
		case "feed":
			return FeedTypeAtom
		default:
			return FeedTypeUnknown
		}
	} else if firstChar == '{' {
		// Check if document is valid JSON
		if jsoniter.Valid(buffer.Bytes()) {
			return FeedTypePenLandJSON
		}
	}
	return FeedTypeUnknown
}

func FindRoot(p *xpp.XMLPullParser) (event xpp.XMLEventType, err error) {
	for {
		event, err = p.Next()
		if err != nil {
			return event, err
		}
		if event == xpp.StartTag {
			break
		}

		if event == xpp.EndDocument {
			return event, errors.New("Failed to find root node before document end.")
		}
	}
	return
}

func NewReaderLabel(label string, input io.Reader) (io.Reader, error) {
	conv, err := charset.NewReaderLabel(label, input)

	if err != nil {
		return nil, err
	}

	// Wrap the charset decoder reader with a XML sanitizer
	//clean := NewXMLSanitizerReader(conv)
	return conv, nil
}
