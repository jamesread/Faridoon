import { createRouter, createWebHistory } from 'vue-router'
import QuotesIndex from '../views/QuotesIndex.vue'
import QuoteShow from '../views/QuoteShow.vue'
import QuoteCreate from '../views/QuoteCreate.vue'
import QuoteEdit from '../views/QuoteEdit.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Account from '../views/Account.vue'
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
  { path: '/account', name: 'account', component: Account },
  {
    path: '/approvals',
    name: 'approvals',
    component: Approvals,
    meta: { requiresApprover: true },
  },
  { path: '/admin/users', name: 'admin-users', component: UsersAdmin, meta: { requiresAdmin: true } },
  {
    path: '/admin/users/:id/edit',
    name: 'admin-user-edit',
    component: UserEdit,
    props: true,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/users/:id',
    name: 'admin-user',
    component: UserDetails,
    props: true,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/groups/create',
    name: 'admin-group-create',
    component: GroupCreate,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/groups/:id',
    name: 'admin-group',
    component: GroupDetails,
    props: true,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/webhooks',
    name: 'admin-webhooks',
    component: WebhooksAdmin,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/webhooks/create',
    name: 'admin-webhook-create',
    component: WebhookCreate,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/header-links',
    name: 'admin-header-links',
    component: HeaderLinksAdmin,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/header-links/create',
    name: 'admin-header-link-create',
    component: HeaderLinkCreate,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/settings',
    name: 'admin-settings',
    component: SettingsAdmin,
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/diagnostics',
    name: 'admin-diagnostics',
    component: DiagnosticsAdmin,
    meta: { requiresAdmin: true },
  },
  { path: '/admin/logs', name: 'admin-logs', component: LogsAdmin, meta: { requiresAdmin: true } },
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

  if (to.meta.requiresAdmin && !user?.isAdmin) {
    return loginRedirect(to)
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
