package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUrl(t *testing.T) {
	urlStr := getUrl()
	assert.True(t, strings.HasPrefix(urlStr, availabilityAPI))
}
