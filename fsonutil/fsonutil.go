// Copyright (c) 2025 Lucas Rocukhout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package fsonutil provides utility tools that are not part of the core fson library
package fsonutil

import (
	"sync"
)

const (
	// DefaultInitialSize is the default initial size hint of a buffer in the Pool. Set at 1KiB.
	DefaultInitialSize = 1024
	// DefaultLowerBound defines the default lower bound for a pool. Set at 64KiB.
	// Buffer's with a capacity lower than this size will always be recycled into the pool.
	DefaultLowerBound = 1 << 16
)

// Pool is an implementation for a buffer pool of []byte slices.
//
// Most people will want to use this instead of a raw sync.Pool
// to avoid memory leaks when using fson.
//
// It avoids pinning large chunks of memory (memory leaks) by using a
// very simple "statistical" approach that will discard sequentially
// under-utilized buffers in the pool, devised by 'dsnet'
// https://github.com/golang/go/issues/27735#issuecomment-739169121
// https://github.com/golang/go/issues/23199
//
// This is particularly useful for a lot of use-cases where fson is involved.
// Most of the time the eventual JSON payload you want to encode is "dynamically sized"
// in the sense that you do not know how big each payload will be. Given this situation
// it's easy for a sudden big payload to grow the size of the underlying buffer used
// by fson. Naively putting this big buffer back into a pool will cause an
// ever-growing chunk of memory to be pinned and never picked up by the garbage collector.
type Pool struct {
	pool       sync.Pool
	lowerBound int // the size under which a buffer will always be recycled
}

// PooledBuffer is a simple wrapper around a byte slice.
//
// You can get the underlying byte slice by calling the Bytes()
// function on a PooledBuffer.
type PooledBuffer struct {
	buf     []byte
	strikes int // number of times the buffer was under-utilized
}

// Bytes returns the underlying byte slice of this PooledBuffer.
func (p *PooledBuffer) Bytes() []byte { return p.buf }

type Option = func(*Pool)

// WithSizeHint will set the default initial size of each buffer (byte slice)
// in the buffer pool. This is handy if you already know, on average, how big
// each of the encoded JSON objects will be.
func WithSizeHint(size int) Option {
	return func(p *Pool) {
		p.pool.New = func() interface{} {
			return &PooledBuffer{
				buf:     make([]byte, size),
				strikes: 0,
			}
		}
	}
}

// WithLowerBound will set the lower bound for this object Pool.
// Buffers with a capacity lower than this is will always be recycled
// into the pool
func WithLowerBound(lowerBound int) Option {
	return func(p *Pool) {
		p.lowerBound = lowerBound
	}
}

// NewPool will create a new object pool.
func NewPool(options ...Option) *Pool {
	pool := &Pool{
		pool: sync.Pool{
			New: func() any {
				return &PooledBuffer{
					buf:     make([]byte, DefaultInitialSize),
					strikes: 0,
				}
			},
		},
		lowerBound: DefaultLowerBound,
	}

	for _, option := range options {
		option(pool)
	}

	return pool
}

// Get will return a PooledBuffer from the Pool ready to be used.
func (p *Pool) Get() *PooledBuffer {
	return p.pool.Get().(*PooledBuffer) //nolint: forcetypeassert
}

// Put will recycle the provided PooledBuffer back into the pool.
func (p *Pool) Put(o *PooledBuffer) {
	// If a buffer is under-utilized enough times sequentially,
	// then it is discarded, ensuring that a single large buffer
	// won't be kept alive by a continuous stream of small usages.
	//
	// But, at the same time a large buffer will be recycled if sufficiently utilized.
	// You can tweak the definitions of utilized, but this implementation uses "less than 50% utilization
	// 4 times in a row" as the definition which is the same heuristic as in dsnet's example.
	//
	// For the constants chosen here, the worst case sequence of utilization would be:
	//
	// 50%, 0%, 0%, 0%, 0%, 50%, 0%, 0%, 0%, 0%, 50%, 0%, 0%, 0%, 0%, ...
	// On average, that's a worst case utilization of 10%, which is far better than the theoretical worst-case of 0%
	// if the code naively recycles buffers without any limits.
	//
	// Advantages of the algorithm:
	//
	// * It's simple; easy to analyze and reason about.
	// * It's fast; only requires basic integer arithmetic.
	// * It doesn't depend on any global state where statistics are concurrently gathered.
	//
	// For usages of all the same approximate size (regardless of how big or small),
	// this approach ensures buffers are always recycled.
	switch {
	case cap(o.buf) <= p.lowerBound: // always recycle buffers smaller than the lowerBound default: 64KiB
		o.strikes = 0
	case cap(o.buf)/2 <= len(o.buf): // at least 50% utilization
		o.strikes = 0
	case o.strikes < 4:
		o.strikes++
	default:
		return // discard the buffer; too large and too often under-utilized
	}

	o.buf = o.buf[:0] // Reset the buffer before putting it back into the pool
	p.pool.Put(o)
}
