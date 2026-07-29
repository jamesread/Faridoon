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
import UserEdit from '../views/UserEdit.vue'
import WebhooksAdmin from '../views/WebhooksAdmin.vue'
import WebhookCreate from '../views/WebhookCreate.vue'
import LogsAdmin from '../views/LogsAdmin.vue'

const routes = [
  { path: '/', name: 'home', redirect: { name: 'quotes' } },
  { path: '/quotes', name: 'quotes', component: QuotesIndex },
  { path: '/quotes/create', name: 'quote-create', component: QuoteCreate },
  { path: '/quotes/:id', name: 'quote-show', component: QuoteShow, props: true },
  { path: '/quotes/:id/edit', name: 'quote-edit', component: QuoteEdit, props: true },
  { path: '/login', name: 'login', component: Login },
  { path: '/register', name: 'register', component: Register },
  { path: '/account', name: 'account', component: Account },
  { path: '/approvals', name: 'approvals', component: Approvals },
  { path: '/admin/users', name: 'admin-users', component: UsersAdmin },
  { path: '/admin/users/:id/edit', name: 'admin-user-edit', component: UserEdit, props: true },
  { path: '/admin/webhooks', name: 'admin-webhooks', component: WebhooksAdmin },
  { path: '/admin/webhooks/create', name: 'admin-webhook-create', component: WebhookCreate },
  { path: '/admin/logs', name: 'admin-logs', component: LogsAdmin },
]

export default createRouter({
  history: createWebHistory(),
  routes,
})
