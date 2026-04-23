import { defineComponent } from 'vue';
import { createRouter, createWebHashHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';

/** Renders nothing; `WatchView` is mounted in `AuthenticatedLayout` so the Twitch iframe stays alive. */
const watchOutletStub = defineComponent({
  name: 'WatchOutletStub',
  setup() {
    return () => null;
  },
});

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('../layouts/AuthenticatedLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'watch',
          component: watchOutletStub,
          meta: { fillMainOutlet: false },
        },
        {
          path: 'settings/rules/new',
          name: 'rule-new',
          component: () => import('../views/RuleEditorView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'settings/rules/:id/edit',
          name: 'rule-edit',
          component: () => import('../views/RuleEditorView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../views/SettingsView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'ai',
          name: 'ai',
          component: () => import('../views/AiChatView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'stats',
          name: 'stats',
          component: () => import('../views/StatsView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'messages',
          name: 'messages',
          component: () => import('../views/MessagesView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'notifications',
          redirect: { name: 'rule-triggers' },
        },
        {
          path: 'rule-triggers',
          name: 'rule-triggers',
          component: () => import('../views/RuleTriggersView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'irc-joined',
          name: 'irc-joined-graph',
          component: () => import('../views/IrcJoinedGraphView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'users/:id',
          name: 'user',
          component: () => import('../views/UserView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('../views/UsersView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'streams',
          name: 'streams',
          component: () => import('../views/StreamsView.vue'),
          meta: { fillMainOutlet: true },
        },
        {
          path: 'streams/:id',
          name: 'stream',
          component: () => import('../views/StreamDetailView.vue'),
          meta: { fillMainOutlet: true },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.bootstrapped) {
    await auth.bootstrap();
  }
  if (to.meta.public) {
    if (auth.isAuthenticated && to.name === 'login') {
      return { name: 'watch' };
    }
    return true;
  }
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } };
  }
  return true;
});

export default router;
