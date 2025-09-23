import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth-store'
import HomePage from '../pages/HomePage.vue'
import LoginPage from '../pages/LoginPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginPage
    },
    {
      path: '/',
      name: 'home',
      component: HomePage,
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to, _, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.token) {
    next('/login')
  } else if (to.path === '/login' && authStore.token) {
    next('/')
  } else {
    next()
  }
})

export default router