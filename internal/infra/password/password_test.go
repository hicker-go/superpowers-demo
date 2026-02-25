package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("secret")
	require.NoError(t, err)
	require.True(t, Verify("secret", hash))
	require.False(t, Verify("wrong", hash))
}
