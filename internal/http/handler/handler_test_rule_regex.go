package handler

import (
	"context"
	"regexp"

	"github.com/rofleksey/dredge/internal/http/gen"
)

func (h *Handler) TestRuleRegex(ctx context.Context, req *gen.TestRuleRegexRequest) (*gen.TestRuleRegexResponse, error) {
	pat := req.GetPattern()
	if req.CaseInsensitive.Or(false) {
		pat = "(?i)" + pat
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		var ce gen.OptNilString
		ce.SetTo(err.Error())

		return &gen.TestRuleRegexResponse{
			Matches:      false,
			CompileError: ce,
		}, nil
	}

	var ce gen.OptNilString
	ce.SetToNull()

	return &gen.TestRuleRegexResponse{
		Matches:      re.MatchString(req.GetSample()),
		CompileError: ce,
	}, nil
}
