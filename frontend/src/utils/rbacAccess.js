const CONTROL_PANEL_PERMISSIONS = [
  'users.view',
  'usergroups.view',
  'rbac.view',
  'system.settings',
]

export const IAM_PERMISSIONS = ['users.view', 'usergroups.view', 'rbac.view']

export function statusFromUser(user) {
  if (!user) {
    return { isLoggedIn: false }
  }
  return {
    isLoggedIn: true,
    rbacIsSuperuser: !!user.isAdmin,
    rbacPermissions: user.privileges || [],
  }
}

export function hasPermission(st, name) {
  return Boolean(st?.rbacIsSuperuser) || (st.rbacPermissions || []).includes(name)
}

export function canAccessControlPanelFromStatus(st) {
  if (!st?.isLoggedIn) return false
  if (st.rbacIsSuperuser) return true
  return CONTROL_PANEL_PERMISSIONS.some((p) => (st.rbacPermissions || []).includes(p))
}

export function canAccessIamFromStatus(st) {
  if (!st?.isLoggedIn) return false
  if (st.rbacIsSuperuser) return true
  return IAM_PERMISSIONS.some((p) => (st.rbacPermissions || []).includes(p))
}

export function canAccessSettingsFromStatus(st) {
  if (!st?.isLoggedIn) return false
  if (st.rbacIsSuperuser) return true
  return (st.rbacPermissions || []).includes('system.settings')
}
