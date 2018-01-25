// Copyright (c) 2018 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package processor

import (
	"github.com/Jeffail/benthos/lib/types"
	"github.com/Jeffail/benthos/lib/util/service/log"
	"github.com/Jeffail/benthos/lib/util/service/metrics"
	"github.com/Jeffail/benthos/lib/util/text"
)

//------------------------------------------------------------------------------

func init() {
	constructors["insert_part"] = typeSpec{
		constructor: NewInsertPart,
		description: `
Insert a new message part at an index. If the specified index is greater than
the length of the existing parts it will be appended to the end.

The index can be negative, and if so the part will be inserted from the end
counting backwards starting from -1. E.g. if index = -1 then the new part will
become the last part of the message, if index = -2 then the new part will be
inserted before the last element, and so on.

This processor will interpolate functions within the 'content' field.`,
	}
}

//------------------------------------------------------------------------------

// InsertPartConfig contains any configuration for the InsertPart processor.
type InsertPartConfig struct {
	Index   int    `json:"index" yaml:"index"`
	Content string `json:"content" yaml:"content"`
}

// NewInsertPartConfig returns a InsertPartConfig with default values.
func NewInsertPartConfig() InsertPartConfig {
	return InsertPartConfig{
		Index:   -1,
		Content: "",
	}
}

//------------------------------------------------------------------------------

// InsertPart is a processor that inserts a new message part at a specific
// index.
type InsertPart struct {
	interpolate bool
	part        []byte

	conf  Config
	log   log.Modular
	stats metrics.Type
}

// NewInsertPart returns a InsertPart processor.
func NewInsertPart(conf Config, log log.Modular, stats metrics.Type) (Type, error) {
	part := []byte(conf.InsertPart.Content)
	interpolate := text.ContainsSpecialVariables(part)
	return &InsertPart{
		part:        part,
		interpolate: interpolate,
		conf:        conf,
		log:         log.NewModule(".processor.insert_part"),
		stats:       stats,
	}, nil
}

//------------------------------------------------------------------------------

// ProcessMessage prepends a new message part to the message.
func (p *InsertPart) ProcessMessage(msg *types.Message) (*types.Message, types.Response, bool) {
	p.stats.Incr("processor.insert_part.count", 1)

	var newPart []byte
	if p.interpolate {
		newPart = text.ReplaceSpecialVariables(p.part)
	} else {
		newPart = p.part
	}

	index := p.conf.InsertPart.Index
	if index < 0 {
		index = len(msg.Parts) + index + 1
		if index < 0 {
			index = 0
		}
	} else if index > len(msg.Parts) {
		index = len(msg.Parts)
	}

	var pre, post [][]byte
	if index > 0 {
		pre = msg.Parts[:index]
	}
	if index < len(msg.Parts) {
		post = msg.Parts[index:]
	}

	newParts := make([][]byte, len(msg.Parts)+1)
	newParts[index] = newPart

	copy(newParts, pre)
	copy(newParts[index+1:], post)

	msg.Parts = newParts

	return msg, nil, true
}

//------------------------------------------------------------------------------