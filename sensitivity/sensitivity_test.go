package sensitivity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateSensitivity(t *testing.T) {
	assert := require.New(t)
	s := CalculateSensitivity(125000, 6, -20)

	assert.Equal(-137, int(s))
}

func TestCalculateLinkBudget(t *testing.T) {
	assert := require.New(t)

	lb := CalculateLinkBudget(125000, 6, -20, 17)
	assert.Equal(154, int(lb))
}
