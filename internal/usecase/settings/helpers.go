package settings

func normalizeTwitchAccountLinkType(accountType string) string {
	if accountType == "bot" {
		return "bot"
	}

	return "main"
}
