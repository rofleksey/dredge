<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import AppShell from '../components/AppShell.vue';
import WatchView from '../views/WatchView.vue';
import { useLiveSocketStore } from '../stores/liveSocket';

const route = useRoute();
const live = useLiveSocketStore();
const { connected: liveConnected, lastError: liveError } = storeToRefs(live);

const fillOutlet = computed(() => Boolean(route.meta.fillMainOutlet));
</script>

<template>
  <AppShell :live-connected="liveConnected" :live-error="liveError">
    <div class="auth-content">
      <!--
        Watch is not under router-view: keep-alive still detaches/hides with display:none,
        which pauses Twitch's iframe. Keep one instance mounted; move off-screen on Settings
        (no display:none) so playback continues.
      -->
      <div
        class="watch-persist"
        :class="{ 'watch-persist--offscreen': route.name !== 'watch' }"
      >
        <WatchView />
      </div>
      <div class="auth-outlet" :class="{ 'auth-outlet--fill': fillOutlet }">
        <router-view />
      </div>
    </div>
  </AppShell>
</template>

<style scoped lang="scss">
.auth-content {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
}

.watch-persist {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;

  &--offscreen {
    position: fixed;
    left: -10000px;
    top: 0;
    width: 640px;
    height: 360px;
    flex: 0 0 0;
    min-height: 0;
    overflow: hidden;
    opacity: 0;
    pointer-events: none;
  }
}

.auth-outlet {
  flex: 0 0 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;

  &--fill {
    flex: 1 1 auto;
    min-height: 0;
  }
}
</style>
