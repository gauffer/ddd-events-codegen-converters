package model

type (
	OAuthScope  string
	OAuthScopes []OAuthScope
)

func (s OAuthScopes) Strings() []string {
	strs := make([]string, len(s))
	for i, scope := range s {
		strs[i] = string(scope)
	}
	return strs
}

func (s OAuthScopes) Equals(other OAuthScopes) bool {
	if len(s) != len(other) {
		return false
	}
	for i := range s {
		if s[i] != other[i] {
			return false
		}
	}
	return true
}

func (s OAuthScopes) ContainsAll(other OAuthScopes) bool {
	if len(other) == 0 {
		return true
	}
	if len(s) == 0 {
		return false
	}

	allowedMap := make(map[OAuthScope]bool, len(s))
	for _, scope := range s {
		allowedMap[scope] = true
	}

	for _, scope := range other {
		if !allowedMap[scope] {
			return false
		}
	}
	return true
}

type OAuthClientID string

type OAuthClientType string

const (
	OAuthClientTypeConfidential OAuthClientType = "confidential"
	OAuthClientTypePublic       OAuthClientType = "public"
)

func (t OAuthClientType) IsValid() bool {
	return t == OAuthClientTypeConfidential || t == OAuthClientTypePublic
}

func (t OAuthClientType) IsConfidential() bool {
	return t == OAuthClientTypeConfidential
}

type OAuthGrantType string

const (
	OAuthGrantTypeAuthorizationCode OAuthGrantType = "authorization_code"
	OAuthGrantTypeClientCredentials OAuthGrantType = "client_credentials"
	OAuthGrantTypeRefreshToken      OAuthGrantType = "refresh_token"
)

func (g OAuthGrantType) IsValid() bool {
	return g == OAuthGrantTypeAuthorizationCode ||
		g == OAuthGrantTypeClientCredentials ||
		g == OAuthGrantTypeRefreshToken
}

type SubjectType string

const (
	SubjectTypeUser    SubjectType = "user"
	SubjectTypeAdmin   SubjectType = "admin"
	SubjectTypeService SubjectType = "service"
)

func (s SubjectType) IsValid() bool {
	return s == SubjectTypeUser || s == SubjectTypeAdmin || s == SubjectTypeService
}

func (s SubjectType) IsService() bool {
	return s == SubjectTypeService
}
