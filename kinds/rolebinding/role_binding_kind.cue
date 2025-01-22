spec: {
	// The role we are discussing
	role: #BuiltinRoleRef | #CustomRoleRef

	// The team or user that has the specified role
	subject: #RoleBindingSubject
} @cuetsy(kind="interface")

#CustomRoleRef: {
	kind: "Role"
	name: string
} @cuetsy(kind="interface")

#BuiltinRoleRef: {
	kind: "BuiltinRole"
	name: "viewer" | "editor" | "admin"
} @cuetsy(kind="interface")

#RoleBindingSubject: {
	kind: "Team" | "User"

	// The team/user identifier name
	name: string
} @cuetsy(kind="interface")
