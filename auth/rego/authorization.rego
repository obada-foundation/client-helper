package obada.rego

default allowAny = false
default allowOnlyUser = false

roleUser := "USER"
roleAll := {roleUser}

allowAny {
	roles_from_claims := {role | role := input.Roles[_]}
	input_role_is_in_claim := roleAll & roles_from_claims
	count(input_role_is_in_claim) > 0
}

allowOnlyUser {
	roles_from_claims := {role | role := input.Roles[_]}
	input_role_is_in_claim := {roleUser} & roles_from_claims
	count(input_role_is_in_claim) > 0
}
