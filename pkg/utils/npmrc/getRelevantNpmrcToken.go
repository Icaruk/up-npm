package npmrc

type NpmrcTokenLevel string

const (
	Project NpmrcTokenLevel = "project"
	User    NpmrcTokenLevel = "user"
	Global  NpmrcTokenLevel = "global"
	Builtin NpmrcTokenLevel = "builtin"
)

// Gets the most relevant npmrc token using order preference from https://docs.npmjs.com/cli/v10/configuring-npm/npmrc#files
func GetRelevantNpmrcToken(npmrcTokens NpmrcTokens) (string, NpmrcTokenLevel) {

	if npmrcTokens.Project != "" {
		return npmrcTokens.Project, Project
	}

	if npmrcTokens.User != "" {
		return npmrcTokens.User, User
	}

	if npmrcTokens.Global != "" {
		return npmrcTokens.Global, Global
	}

	if npmrcTokens.Builtin != "" {
		return npmrcTokens.Builtin, Builtin
	}

	return "", ""

}
