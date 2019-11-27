package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	s := &Server{}
	c := []CarparkAvailability{{AvailableLots: 10}}

	_, ok := s.getCache()
	assert.False(t, ok)

	s.updateCache(c)
	res, ok := s.getCache()
	assert.True(t, ok)
	assert.Equal(t, c, res)

	s.invalidateCache()
	_, ok = s.getCache()
	assert.False(t, ok)
}
