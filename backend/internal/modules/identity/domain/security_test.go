package domain

import "testing"

func TestRolePermissionBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		role    Role
		allowed Permission
		denied  Permission
	}{
		{"catalog editor", RoleCatalogEditor, PermissionCatalogUpdate, PermissionEvidenceApprove},
		{"evidence reviewer", RoleEvidenceReviewer, PermissionEvidenceApprove, PermissionPolicyActivate},
		{"commerce operator", RoleCommerceOperator, PermissionCommerceActivate, PermissionPolicyCreate},
		{"content editor", RoleContentEditor, PermissionContentUpdate, PermissionCommerceUpdate},
		{"analyst", RoleAnalyst, PermissionAnalyticsExport, PermissionCommerceUpdate},
		{"policy editor", RolePolicyEditor, PermissionPolicyCreate, PermissionPolicyApprove},
		{"policy reviewer", RolePolicyReviewer, PermissionPolicyActivate, PermissionPolicyCreate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			principal := Principal{Roles: []Role{test.role}}
			if !principal.HasPermission(test.allowed) {
				t.Fatalf("%s lacks %s", test.role, test.allowed)
			}
			if principal.HasPermission(test.denied) {
				t.Fatalf("%s unexpectedly has %s", test.role, test.denied)
			}
		})
	}
	if !(&Principal{Roles: []Role{RoleAdmin}}).HasPermission(PermissionPolicyActivate) {
		t.Fatal("administrator must retain all permissions")
	}
}
