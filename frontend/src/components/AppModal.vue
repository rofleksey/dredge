<script setup lang="ts">
import { onUnmounted, watch } from 'vue';

const props = withDefaults(
  defineProps<{
    open: boolean;
    title?: string;
    /** Wider panel for dense content (e.g. stream details). */
    wide?: boolean;
    /** Largest breakpoint (e.g. stream details + chatter list). */
    extraWide?: boolean;
  }>(),
  { title: '', wide: false, extraWide: false },
);

const emit = defineEmits<{ close: [] }>();

function onDocKeydown(e: KeyboardEvent): void {
  if (e.key === 'Escape') {
    emit('close');
  }
}

watch(
  () => props.open,
  (v) => {
    if (v) {
      document.addEventListener('keydown', onDocKeydown);
    } else {
      document.removeEventListener('keydown', onDocKeydown);
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  document.removeEventListener('keydown', onDocKeydown);
});
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="modal-root">
      <div class="modal-backdrop" aria-hidden="true" @click="emit('close')" />
      <div
        class="modal-panel"
        :class="{ 'modal-panel--wide': wide, 'modal-panel--extra-wide': extraWide }"
        role="dialog"
        :aria-label="title || 'Dialog'"
      >
        <header class="modal-head">
          <h2 class="modal-title">
            <slot name="title">{{ title }}</slot>
          </h2>
          <button type="button" class="btn-close" aria-label="Close" @click="emit('close')">×</button>
        </header>
        <div class="modal-body">
          <slot />
        </div>
        <footer v-if="$slots.footer" class="modal-foot">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.modal-root {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  box-sizing: border-box;
}

.modal-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
}

.modal-panel {
  position: relative;
  z-index: 1;
  width: min(100%, 28rem);
  max-height: min(90vh, 32rem);
  display: flex;
  flex-direction: column;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.4rem;
  box-shadow: 0 0.5rem 2rem rgba(0, 0, 0, 0.35);

  &--wide {
    width: min(100%, 46rem);
    max-height: min(92vh, 42rem);
  }

  &--extra-wide {
    width: min(100%, min(96vw, 58rem));
    max-height: min(94vh, 52rem);
  }
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.3;
}

.btn-close {
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  border: none;
  border-radius: 0.25rem;
  background: transparent;
  color: var(--text-muted);
  font-size: 1.35rem;
  line-height: 1;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
    color: var(--text);
  }
}

.modal-body {
  padding: 0.75rem;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
  font-size: 0.88rem;
}

.modal-foot {
  padding: 0.65rem 0.75rem;
  border-top: 1px solid var(--border);
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: flex-end;
  flex-shrink: 0;
}
</style>
