/**
 * User role constants
 * These should match the backend role definitions
 */

export const USER_ROLES = {
  USER: 'user',
  MODERATOR: 'moderator',
  ADMIN: 'admin',
} as const;

export type UserRole = typeof USER_ROLES[keyof typeof USER_ROLES];

/**
 * Check if a role string is a valid user role
 */
export function isValidRole(role: string): role is UserRole {
  return Object.values(USER_ROLES).includes(role as UserRole);
}

/**
 * Check if a user has a specific role
 */
export function hasRole(userRole: UserRole | undefined, requiredRole: UserRole): boolean {
  return userRole === requiredRole;
}

/**
 * Check if a user has any of the specified roles
 */
export function hasAnyRole(userRole: UserRole | undefined, requiredRoles: UserRole[]): boolean {
  if (!userRole) return false;
  return requiredRoles.includes(userRole);
}

/**
 * Check if a user is an admin
 */
export function isAdmin(role: UserRole | undefined): boolean {
  return role === USER_ROLES.ADMIN;
}

/**
 * Check if a user is a moderator
 */
export function isModerator(role: UserRole | undefined): boolean {
  return role === USER_ROLES.MODERATOR;
}

/**
 * Check if a user is a moderator or admin
 */
export function isModeratorOrAdmin(role: UserRole | undefined): boolean {
  return role === USER_ROLES.MODERATOR || role === USER_ROLES.ADMIN;
}

/**
 * Get a display name for a role
 */
export function getRoleDisplayName(role: UserRole): string {
  switch (role) {
    case USER_ROLES.ADMIN:
      return 'Admin';
    case USER_ROLES.MODERATOR:
      return 'Moderator';
    case USER_ROLES.USER:
      return 'User';
    default:
      return 'Unknown';
  }
}

/**
 * Get a role badge color
 */
export function getRoleBadgeColor(role: UserRole): string {
  switch (role) {
    case USER_ROLES.ADMIN:
      return 'bg-error-600 text-white ring-1 ring-error-500/30 font-heading text-[11px] uppercase tracking-wide';
    case USER_ROLES.MODERATOR:
      return 'bg-warning-600 text-neutral-900 ring-1 ring-warning-500/30 font-heading text-[11px] uppercase tracking-wide';
    case USER_ROLES.USER:
      return 'bg-muted text-muted-foreground';
    default:
      return 'bg-muted text-muted-foreground';
  }
}

/**
 * Get a badge variant for a role
 * Used by the UserRoleBadge component
 */
export function getRoleBadgeVariant(role: UserRole): 'error' | 'warning' | 'default' {
  switch (role) {
    case USER_ROLES.ADMIN:
      return 'error';
    case USER_ROLES.MODERATOR:
      return 'warning';
    case USER_ROLES.USER:
    default:
      return 'default';
  }
}
