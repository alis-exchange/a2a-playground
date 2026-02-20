import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/pages/playground/PlaygroundView.vue'),
    },
    {
      path: '/oauth-callback',
      component: () => import('@/pages/OAuthCallbackView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

// When OAuth redirects to root (or any route) with ?code=, ensure we land on /oauth-callback
// so OAuthCallbackView runs and can postMessage + close the popup.
router.beforeEach((to, _from, next) => {
  if (to.path === '/oauth-callback') return next()
  const code = to.query?.code
  if (code) {
    next({ path: '/oauth-callback', query: to.query })
  } else {
    next()
  }
})

export default router
