package model

// 角色常量定义
const (
	RoleSuperAdmin  = "super_admin"
	RoleAdmin       = "admin"
	RoleAuthor      = "author"
	RoleContributor = "contributor"
	RoleGuest       = "guest"
)

// IsSuperAdmin 检查是否为超级管理员
func IsSuperAdmin(user User) bool {
	return user.Role == RoleSuperAdmin
}

// IsAdmin 检查是否为管理员（包括超级管理员）
func IsAdmin(user User) bool {
	return user.Role == RoleAdmin || user.Role == RoleSuperAdmin
}

// IsAuthor 检查是否为作者（包括管理员和超级管理员）
func IsAuthor(user User) bool {
	return user.Role == RoleAuthor || user.Role == RoleAdmin || user.Role == RoleSuperAdmin
}

// IsContributor 检查是否为投稿者（包括更高权限的角色）
func IsContributor(user User) bool {
	return user.Role == RoleContributor || user.Role == RoleAuthor ||
		user.Role == RoleAdmin || user.Role == RoleSuperAdmin
}
