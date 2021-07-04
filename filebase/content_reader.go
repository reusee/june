// Copyright 2021 The June Authors. All rights reserved.
// Use of this source code is governed by Apache License
// that can be found in the LICENSE file.

package filebase

import (
	"fmt"
	"io"

	"github.com/reusee/e4"
	"github.com/reusee/june/entity"
)

type ContentReader struct {
	fetch entity.Fetch

	remainBytes []byte
	keys        []Key
	keyIndex    int
	offset      int64

	lengths []*int64
}

var _ io.Reader = new(ContentReader)

type NewContentReader func(
	keys []Key,
	lengths []int64,
) *ContentReader

func (_ Def) NewContentReader(
	fetch entity.Fetch,
) NewContentReader {
	return func(
		keys []Key,
		lengths []int64,
	) *ContentReader {
		r := &ContentReader{
			fetch:   fetch,
			keys:    keys,
			lengths: make([]*int64, len(keys)),
		}
		if len(lengths) == len(keys) {
			for i := range lengths {
				r.lengths[i] = &lengths[i]
			}
		}
		return r
	}
}

func (c *ContentReader) Read(buf []byte) (_ int, err error) {
	defer he(&err)

	if len(c.remainBytes) > 0 {
		n := copy(buf, c.remainBytes)
		c.remainBytes = c.remainBytes[n:]
		c.offset += int64(n)
		return n, nil
	}

	if c.keyIndex >= len(c.keys) {
		return 0, io.EOF
	}

	key := c.keys[c.keyIndex]
	var content Content
	ce(c.fetch(key, &content),
		e4.NewInfo("fetch %s", key))
	c.remainBytes = content
	l := int64(len(content))
	c.lengths[c.keyIndex] = &l
	c.keyIndex++
	return c.Read(buf)
}

func (c *ContentReader) getLen() (ret int64) {
	for i, p := range c.lengths {
		if p == nil {
			key := c.keys[i]
			var content Content
			ce(c.fetch(key, &content),
				e4.NewInfo("fetch %s", key))
			l := int64(len(content))
			c.lengths[i] = &l
			ret += l
		} else {
			ret += *p
		}
	}
	return ret
}

func (c *ContentReader) Seek(offset int64, whence int) (_ int64, err error) {
	defer he(&err)

	if whence == io.SeekCurrent {
		return c.Seek(c.offset+offset, io.SeekStart)
	} else if whence == io.SeekEnd {
		return c.Seek(c.getLen()+offset, io.SeekStart)
	}

	if offset < 0 {
		return 0, fmt.Errorf("bad offset: %d", offset)
	}

	c.remainBytes = c.remainBytes[:0]
	c.keyIndex = 0
	c.offset = 0
	for i, p := range c.lengths {
		c.keyIndex = i + 1

		if p == nil {
			key := c.keys[i]
			var content Content
			ce(c.fetch(key, &content),
				e4.NewInfo("fetch %s", key))
			l := int64(len(content))
			c.lengths[i] = &l
			if offset <= c.offset+l {
				cut := offset - c.offset
				c.remainBytes = content[cut:]
				c.offset = offset
				return c.offset, nil
			} else {
				c.offset += l
				continue
			}

		} else {
			l := *p
			if offset <= c.offset+l {
				key := c.keys[i]
				var content Content
				ce(c.fetch(key, &content),
					e4.NewInfo("fetch %s", key))
				cut := offset - c.offset
				c.remainBytes = content[cut:]
				c.offset = offset
				return c.offset, nil
			} else {
				c.offset += l
				continue
			}
		}
	}

	return c.offset, nil
}
