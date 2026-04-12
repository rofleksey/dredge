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
                },
                {
                    path: 'settings',
                    name: 'settings',
                    component: () => import('../views/SettingsView.vue'),
                },
                {
                    path: 'messages',
                    name: 'messages',
                    component: () => import('../views/MessagesView.vue'),
                },
                {
                    path: 'users/:id',
                    name: 'user',
                    component: () => import('../views/UserView.vue'),
                },
                {
                    path: 'users',
                    name: 'users',
                    component: () => import('../views/UsersView.vue'),
                },
                {
                    path: 'streams',
                    name: 'streams',
                    component: () => import('../views/StreamsView.vue'),
                },
                {
                    path: 'streams/:id',
                    name: 'stream',
                    component: () => import('../views/StreamDetailView.vue'),
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
