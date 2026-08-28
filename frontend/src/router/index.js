import { createRouter, createWebHistory } from 'vue-router'
import QuotesIndex from '../views/QuotesIndex.vue'
import QuoteShow from '../views/QuoteShow.vue'
import QuoteCreate from '../views/QuoteCreate.vue'
import QuoteEdit from '../views/QuoteEdit.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import UserControlPanel from '../views/UserControlPanel.vue'
import UserPreferences from '../views/UserPreferences.vue'
import ChangePassword from '../views/ChangePassword.vue'
import MyPermissions from '../views/MyPermissions.vue'
import ControlPanel from '../views/ControlPanel.vue'
import Iam from '../views/Iam.vue'
import Approvals from '../views/Approvals.vue'
import UsersAdmin from '../views/UsersAdmin.vue'
import UserDetails from '../views/UserDetails.vue'
import UserEdit from '../views/UserEdit.vue'
import GroupDetails from '../views/GroupDetails.vue'
import GroupCreate from '../views/GroupCreate.vue'
import WebhooksAdmin from '../views/WebhooksAdmin.vue'
import WebhookCreate from '../views/WebhookCreate.vue'
import LogsAdmin from '../views/LogsAdmin.vue'
import HeaderLinksAdmin from '../views/HeaderLinksAdmin.vue'
import HeaderLinkCreate from '../views/HeaderLinkCreate.vue'
import SettingsAdmin from '../views/SettingsAdmin.vue'
import DiagnosticsAdmin from '../views/DiagnosticsAdmin.vue'
import { initState, loadInit } from '../composables/useInit'
import {
  canAccessControlPanelFromStatus,
  canAccessIamFromStatus,
  canAccessSettingsFromStatus,
  statusFromUser,
} from '../utils/rbacAccess'

const routes = [
  { path: '/', name: 'home', redirect: { name: 'quotes' } },
  { path: '/quotes', name: 'quotes', component: QuotesIndex },
  {
    path: '/quotes/create',
    name: 'quote-create',
    component: QuoteCreate,
    meta: { requiresAuthOrGuestAdd: true },
  },
  { path: '/quotes/:id', name: 'quote-show', component: QuoteShow, props: true },
  {
    path: '/quotes/:id/edit',
    name: 'quote-edit',
    component: QuoteEdit,
    props: true,
    meta: { requiresApprover: true },
  },
  { path: '/login', name: 'login', component: Login },
  {
    path: '/register',
    name: 'register',
    component: Register,
    meta: { requiresRegistration: true },
  },
  { path: '/account', redirect: { name: 'userControlPanel' } },
  {
    path: '/user-control-panel',
    name: 'userControlPanel',
    component: UserControlPanel,
    meta: { title: 'User Control Panel', requiresAuth: true },
  },
  {
    path: '/user-control-panel/preferences',
    name: 'userPreferences',
    component: UserPreferences,
    meta: { title: 'User Preferences', requiresAuth: true },
  },
  {
    path: '/user-control-panel/permissions',
    name: 'myPermissions',
    component: MyPermissions,
    meta: { title: 'My Permissions', requiresAuth: true },
  },
  {
    path: '/change-password',
    name: 'changePassword',
    component: ChangePassword,
    meta: { title: 'Change Password', requiresAuth: true },
  },
  {
    path: '/control-panel',
    name: 'controlPanel',
    component: ControlPanel,
    meta: { title: 'System Control Panel', requiresAuth: true, requiresControlPanel: true },
  },
  {
    path: '/control-panel/iam',
    name: 'iam',
    component: Iam,
    meta: { title: 'IAM', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/iam/users',
    name: 'iamUsers',
    component: UsersAdmin,
    meta: { title: 'Users', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/iam/users/:id/edit',
    name: 'iamUserEdit',
    component: UserEdit,
    props: true,
    meta: { title: 'Edit user', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/iam/users/:id',
    name: 'iamUser',
    component: UserDetails,
    props: true,
    meta: { title: 'User', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/iam/groups/create',
    name: 'iamGroupCreate',
    component: GroupCreate,
    meta: { title: 'Create group', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/iam/groups/:id',
    name: 'iamGroup',
    component: GroupDetails,
    props: true,
    meta: { title: 'Group', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/settings',
    name: 'settings',
    component: SettingsAdmin,
    meta: { title: 'Settings', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/webhooks',
    name: 'webhooks',
    component: WebhooksAdmin,
    meta: { title: 'Webhooks', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/webhooks/create',
    name: 'webhookCreate',
    component: WebhookCreate,
    meta: { title: 'Add webhook', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/header-links',
    name: 'headerLinks',
    component: HeaderLinksAdmin,
    meta: { title: 'Header links', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/header-links/create',
    name: 'headerLinkCreate',
    component: HeaderLinkCreate,
    meta: { title: 'Add header link', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/diagnostics',
    name: 'diagnostics',
    component: DiagnosticsAdmin,
    meta: { title: 'Diagnostics', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/logs',
    name: 'logs',
    component: LogsAdmin,
    meta: { title: 'Audit logs', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/approvals',
    name: 'approvals',
    component: Approvals,
    meta: { requiresApprover: true },
  },
  { path: '/admin/users', redirect: { name: 'iamUsers' } },
  { path: '/admin/users/:id/edit', redirect: (to) => ({ name: 'iamUserEdit', params: to.params }) },
  { path: '/admin/users/:id', redirect: (to) => ({ name: 'iamUser', params: to.params }) },
  { path: '/admin/groups/create', redirect: { name: 'iamGroupCreate' } },
  { path: '/admin/groups/:id', redirect: (to) => ({ name: 'iamGroup', params: to.params }) },
  { path: '/admin/webhooks', redirect: { name: 'webhooks' } },
  { path: '/admin/webhooks/create', redirect: { name: 'webhookCreate' } },
  { path: '/admin/header-links', redirect: { name: 'headerLinks' } },
  { path: '/admin/header-links/create', redirect: { name: 'headerLinkCreate' } },
  { path: '/admin/settings', redirect: { name: 'settings' } },
  { path: '/admin/diagnostics', redirect: { name: 'diagnostics' } },
  { path: '/admin/logs', redirect: { name: 'logs' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

function loginRedirect(to) {
  return { name: 'login', query: { redirect: to.fullPath } }
}

async function ensureInit() {
  if (!initState.ready) {
    await loadInit()
  }
}

router.beforeEach(async (to) => {
  await ensureInit()
  const user = initState.user
  const features = initState.features
  const st = statusFromUser(user)

  if (to.meta.requiresAuth && !user) {
    return loginRedirect(to)
  }
  if (to.meta.requiresControlPanel && !canAccessControlPanelFromStatus(st)) {
    return { name: 'quotes' }
  }
  if (to.meta.requiresIam && !canAccessIamFromStatus(st)) {
    return { name: 'quotes' }
  }
  if (to.meta.requiresSettings && !canAccessSettingsFromStatus(st)) {
    return { name: 'quotes' }
  }
  if (to.meta.requiresApprover && !user?.canApproveQuotes) {
    return loginRedirect(to)
  }
  if (to.meta.requiresRegistration && !features.registrationEnabled) {
    return { name: 'quotes' }
  }
  if (to.meta.requiresAuthOrGuestAdd && !user && !features.guestAddEnabled) {
    return loginRedirect(to)
  }
  return true
})

export default router
