import { createRouter, createWebHistory, type RouterHistory } from 'vue-router'

export const routes = [
  { path: '/', name: 'list', component: () => import('@/views/ClaimListView.vue') },
  { path: '/new', name: 'new', component: () => import('@/views/ClaimFormView.vue') },
  { path: '/claims/:id', name: 'detail', component: () => import('@/views/ClaimDetailView.vue'), props: true },
  { path: '/:rest(.*)*', name: 'not-found', component: () => import('@/views/NotFoundView.vue') },
]

export function makeRouter(history: RouterHistory = createWebHistory()) {
  return createRouter({ history, routes })
}
