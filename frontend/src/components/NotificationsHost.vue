<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { watch } from 'vue';
import { notify } from '../lib/notify';
import { useLiveSocketStore } from '../stores/liveSocket';
import { useNotificationsStore } from '../stores/notifications';

const notifications = useNotificationsStore();
const { items } = storeToRefs(notifications);

const live = useLiveSocketStore();
const { lastError } = storeToRefs(live);

watch(
  lastError,
  (msg) => {
    if (msg) {
      notify({
        id: 'live-ws',
        type: 'warning',
        title: 'Live connection',
        description: msg,
      });
    }
  },
  { immediate: true },
);
</script>

<template>
  <Teleport to="body">
    <div class="notifications-host" aria-live="polite">
      <div
        v-for="n in items"
        :key="n.key"
        class="toast"
        :class="`toast--${n.type}`"
        role="status"
      >
        <div class="toast-body">
          <p class="toast-title">{{ n.title }}</p>
          <p class="toast-desc">{{ n.description }}</p>
        </div>
        <div class="toast-progress-wrap" aria-hidden="true">
          <div
            class="toast-progress"
            :style="{ animationDuration: `${n.durationMs}ms` }"
          />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
@keyframes toast-progress-shrink {
  from {
    transform: scaleX(1);
  }
  to {
    transform: scaleX(0);
  }
}

.notifications-host {
  position: fixed;
  top: 0.75rem;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3000;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  pointer-events: none;
  max-width: min(420px, calc(100vw - 1.5rem));
}

.toast {
  width: 100%;
  border-radius: 0.4rem;
  border: 1px solid var(--border, #2f2f35);
  background: var(--bg-elevated, #18181b);
  color: var(--text, #efeff1);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
  overflow: hidden;
  pointer-events: auto;
}

.toast-body {
  padding: 0.65rem 0.85rem 0.5rem;
}

.toast-title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  line-height: 1.3;
}

.toast-desc {
  margin: 0.25rem 0 0;
  font-size: 0.82rem;
  color: var(--text-muted, #adadb8);
  line-height: 1.4;
}

.toast-progress-wrap {
  height: 3px;
  background: rgba(255, 255, 255, 0.06);
}

.toast-progress {
  height: 100%;
  width: 100%;
  transform-origin: left center;
  animation-name: toast-progress-shrink;
  animation-timing-function: linear;
  animation-fill-mode: forwards;
}

.toast--success .toast-progress {
  background: #00f593;
}

.toast--info .toast-progress {
  background: var(--accent-bright, #bf94ff);
}

.toast--warning .toast-progress {
  background: #e6a23c;
}

.toast--error .toast-progress {
  background: #ff6b6b;
}

.toast--success {
  border-color: rgba(0, 245, 147, 0.35);
}

.toast--info {
  border-color: rgba(191, 148, 255, 0.35);
}

.toast--warning {
  border-color: rgba(230, 162, 60, 0.45);
}

.toast--error {
  border-color: rgba(255, 107, 107, 0.45);
}
</style>
