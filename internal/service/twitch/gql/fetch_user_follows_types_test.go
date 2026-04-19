package gql

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGqlFollowResponse_unmarshal(t *testing.T) {
	t.Parallel()

	raw := `{
  "data": {
    "user": {
      "follows": {
        "totalCount": 3,
        "pageInfo": { "hasNextPage": false, "endCursor": "" },
        "edges": [
          {
            "followedAt": "2021-06-15T12:30:00Z",
            "node": { "id": "987654321", "login": "SomeChannel" }
          }
        ]
      }
    }
  }
}`

	var parsed gqlFollowResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &parsed))
	require.NotNil(t, parsed.Data.User)
	assert.Equal(t, 3, parsed.Data.User.Follows.TotalCount)
	require.Len(t, parsed.Data.User.Follows.Edges, 1)
	edge := parsed.Data.User.Follows.Edges[0]
	assert.Equal(t, "987654321", edge.Node.ID)
	assert.Equal(t, "SomeChannel", edge.Node.Login)

	chID, err := parseTwitchUserID(edge.Node.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(987654321), chID)

	ft, err := time.Parse(time.RFC3339, edge.FollowedAt)
	require.NoError(t, err)
	assert.Equal(t, 2021, ft.UTC().Year())
}
