package list

import (
	"bufio"
	"errors"
	"io"
)

const (
	MdEncCustom = "markdown-custom"
	// https://spec.commonmark.org/0.31.2/#lists
	MdEncCommonMark = "markdown-commonmark"
)

type Markdown struct {
	Reader MarkdownReader
	Writer MarkdownWriter
}

type MarkdownReader struct {
	io.Reader
}

func (mr *MarkdownReader) NewGroupMarkdownReader(mdEnc string, reader io.Reader) ([]*Group, error) {
	switch mdEnc {
	case MdEncCustom:
		lineReader := bufio.NewReader(reader)
		list := make([]*Group, 0)
		var line string
		var err error
		var currentGroup *Group = nil
		for {
			line, err = lineReader.ReadString('\n')
			lineRunes := []rune(line)
			if err != nil {
				break
			}
			if isGroupStart(lineRunes) {
				if currentGroup != nil {
					list = append(list, currentGroup)
				}
				currentGroup = &Group{Items: make([]*Item, 0), Name: string(lineRunes[2:])}
			} else {
				if currentGroup == nil {
					currentGroup = &Group{Name: "", Items: make([]*Item, 0)}
				}
				currentGroup.Items = append(currentGroup.Items, &Item{})
			}
		}
		if currentGroup != nil {
			list = append(list, currentGroup)
		}
		return list, nil
	case MdEncCommonMark:
		panic("CommonMark encoding not supported")
	default:
		return nil, errors.New("unknown markdown encoding")
	}
}

func isGroupStart(lineRunes []rune) bool {
	var listTypes []rune = []rune{'-', '+', '*'}
	runeIndex := 0
	for runeIndex, rune := range lineRunes {
		if rune == ' ' {
			runeIndex++
		} else {
			break
		}
	}
	return contains(listTypes, lineRunes[runeIndex])
}

func contains(runes []rune, r rune) bool {
	for _, rr := range runes {
		if rr == r {
			return true
		}
	}
	return false
}

type MarkdownWriter struct {
	io.Writer
}

func (mw *MarkdownWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}
