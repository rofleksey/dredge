<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';

const route = useRoute();
const router = useRouter();

const props = withDefaults(
  defineProps<{
    liveConnected?: boolean;
    liveError?: string | null;
  }>(),
  {
    liveConnected: false,
    liveError: null,
  },
);

const auth = useAuthStore();

const brandLogoSrc = `${import.meta.env.BASE_URL}logo.jpg`;

const liveTitle = computed(() => {
  if (props.liveConnected) {
    return 'Live: connected';
  }
  if (props.liveError) {
    return `Live: connection problem (${props.liveError})`;
  }
  return 'Live: connecting…';
});

function logout(): void {
  auth.logout();
  void router.push({ name: 'login' });
}
</script>

<template>
  <div class="shell">
    <header class="top">
      <RouterLink class="brand" to="/" aria-label="Dredge home">
        <img class="brand-logo" :src="brandLogoSrc" alt="" width="120" height="32" />
      </RouterLink>
      <nav class="nav">
        <RouterLink to="/" active-class="active">Watch</RouterLink>
        <RouterLink to="/messages" active-class="active">Messages</RouterLink>
        <RouterLink
          to="/users"
          :class="{ active: route.name === 'users' || route.name === 'user' }"
        >
          Users
        </RouterLink>
        <RouterLink
          to="/streams"
          :class="{ active: route.name === 'streams' || route.name === 'stream' }"
        >
          Streams
        </RouterLink>
        <RouterLink to="/ai" active-class="active">AI</RouterLink>
        <RouterLink to="/settings" active-class="active">Settings</RouterLink>
      </nav>
      <div class="user">
        <RouterLink
          to="/notifications"
          class="icon-link"
          :class="{ active: route.name === 'notifications' }"
          aria-label="Notifications"
          title="Notifications"
        >
          &#128276;
        </RouterLink>
        <span
          class="live-dot"
          :class="{
            'live-dot--ok': liveConnected,
            'live-dot--err': !liveConnected && liveError,
            'live-dot--pending': !liveConnected && !liveError,
          }"
          role="status"
          :aria-label="liveTitle"
          :title="liveTitle"
        />
        <button type="button" class="btn-ghost" @click="logout">Log out</button>
      </div>
    </header>
    <main class="main">
      <slot />
    </main>
  </div>
</template>

<style scoped lang="scss">
.shell {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
  color: var(--text);
}

.top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  padding: 0.5rem 0.75rem;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.brand {
  display: flex;
  align-items: center;
  line-height: 0;
  color: var(--accent);
  text-decoration: none;
}

.brand-logo {
  display: block;
  height: 1.35rem;
  width: auto;
  max-width: 7rem;
  object-fit: contain;
  object-position: left center;
}

.nav {
  display: flex;
  gap: 0.5rem;
  flex: 1;

  a {
    color: var(--text-muted);
    text-decoration: none;
    padding: 0.35rem 0.6rem;
    border-radius: 0.25rem;
    font-size: 0.9rem;

    &.active,
    &:hover {
      color: var(--text);
      background: var(--bg-hover);
    }
  }
}

.user {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-left: auto;
}

.icon-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.8rem;
  height: 1.8rem;
  border-radius: 0.35rem;
  border: 1px solid var(--border);
  color: var(--text-muted);
  text-decoration: none;
  font-size: 1rem;
  line-height: 1;

  &:hover,
  &.active {
    color: var(--text);
    background: var(--bg-hover);
  }
}

.live-dot {
  flex-shrink: 0;
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  background: var(--text-muted);
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.06);

  &.live-dot--ok {
    background: #00f593;
    box-shadow: 0 0 0 2px rgba(0, 245, 147, 0.25);
  }

  &.live-dot--err {
    background: #ff6b6b;
    box-shadow: 0 0 0 2px rgba(255, 107, 107, 0.3);
  }

  &.live-dot--pending {
    background: var(--text-muted);
  }
}

.btn-ghost {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text);
  padding: 0.35rem 0.65rem;
  border-radius: 0.25rem;
  font-size: 0.85rem;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
  }
}

.main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}
</style>
