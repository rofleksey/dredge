package httptransport

import (
	"github.com/rofleksey/dredge/internal/transport/http/gen"
)

func genClientNotice(sev gen.ClientNoticeSeverity, code, msg string) gen.ClientNotice {
	return gen.ClientNotice{
		Severity: sev,
		Code:     code,
		Message:  msg,
	}
}

func genClientNoticeErr(code, msg string) gen.ClientNotice {
	return genClientNotice(gen.ClientNoticeSeverityError, code, msg)
}

func genClientNoticeWarn(code, msg string) gen.ClientNotice {
	return genClientNotice(gen.ClientNoticeSeverityWarning, code, msg)
}
