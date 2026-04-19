<script setup lang="ts">
import { nextTick, onUnmounted, ref, watch } from 'vue';

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

const panelRef = ref<HTMLElement | null>(null);
let previousFocus: HTMLElement | null = null;

const FOCUSABLE_SELECTOR =
  'button:not([disabled]), [href], input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

function focusableIn(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (el) => el.offsetParent !== null || el === document.activeElement,
  );
}

function focusFirstInPanel(): void {
  const panel = panelRef.value;
  if (!panel) {
    return;
  }
  const list = focusableIn(panel);
  if (list.length) {
    list[0].focus();
  } else {
    if (!panel.hasAttribute('tabindex')) {
      panel.setAttribute('tabindex', '-1');
    }
    panel.focus();
  }
}

function onPanelKeydown(e: KeyboardEvent): void {
  if (e.key !== 'Tab' || !panelRef.value) {
    return;
  }
  const list = focusableIn(panelRef.value);
  if (list.length < 2) {
    return;
  }
  const first = list[0];
  const last = list[list.length - 1];
  const active = document.activeElement as HTMLElement | null;
  if (e.shiftKey) {
    if (active === first) {
      e.preventDefault();
      last.focus();
    }
  } else if (active === last) {
    e.preventDefault();
    first.focus();
  }
}

function onDocKeydown(e: KeyboardEvent): void {
  if (e.key === 'Escape') {
    emit('close');
  }
}

watch(
  () => props.open,
  async (v) => {
    if (v) {
      previousFocus = document.activeElement as HTMLElement | null;
      document.addEventListener('keydown', onDocKeydown);
      await nextTick();
      focusFirstInPanel();
    } else {
      document.removeEventListener('keydown', onDocKeydown);
      previousFocus?.focus?.();
      previousFocus = null;
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
        ref="panelRef"
        class="modal-panel"
        :class="{ 'modal-panel--wide': wide, 'modal-panel--extra-wide': extraWide }"
        role="dialog"
        aria-modal="true"
        :aria-label="title || 'Dialog'"
        @keydown="onPanelKeydown"
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
