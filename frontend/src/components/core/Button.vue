<script setup lang="ts">
import { computed } from 'vue';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'outline';
type ButtonSize = 'default' | 'small' | 'tiny';

/**
 * Shared button primitive for action controls.
 * Supports variant/size composition and optional loading spinner.
 */
const props = withDefaults(
  defineProps<{
    variant?: ButtonVariant;
    size?: ButtonSize;
    nativeType?: 'button' | 'submit' | 'reset';
    loading?: boolean;
    disabled?: boolean;
    fullWidth?: boolean;
    inlineSquare?: boolean;
  }>(),
  {
    variant: 'primary',
    size: 'default',
    nativeType: 'button',
    loading: false,
    disabled: false,
    fullWidth: false,
    inlineSquare: false,
  },
);

const classList = computed(() => {
  const list = ['c-btn', `c-btn--${props.variant}`, `c-btn--size-${props.size}`];
  if (props.fullWidth) {
    list.push('c-btn--full');
  }
  if (props.inlineSquare) {
    list.push('c-btn--square');
  }
  return list;
});
</script>

<template>
  <button :type="nativeType" :disabled="loading || disabled" :aria-busy="loading" :class="classList">
    <span v-if="loading" class="btn-submit-spinner" aria-hidden="true" />
    <span><slot /></span>
  </button>
</template>

<style scoped lang="scss">
.c-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  border-radius: 0.25rem;
  font-weight: 600;
  font-size: 0.85rem;
  line-height: 1.2;
  cursor: pointer;
  font-family: inherit;
}

.c-btn--size-default {
  padding: 0.45rem 0.75rem;
}

.c-btn--size-small {
  padding: 0.4rem 0.75rem;
}

.c-btn--size-tiny {
  padding: 0.15rem 0.45rem;
  font-size: 0.75rem;
}

.c-btn--primary {
  border: none;
  background: var(--accent);
  color: #fff;
}

.c-btn--secondary {
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
}

.c-btn--ghost {
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text);
}

.c-btn--danger {
  border: 1px solid var(--border);
  background: transparent;
  color: #ff6b6b;
}

.c-btn--outline {
  border: 1px solid var(--accent);
  background: transparent;
  color: var(--accent-bright);
}

.c-btn--primary:hover:not(:disabled) {
  filter: brightness(1.06);
}

.c-btn--secondary:hover:not(:disabled),
.c-btn--ghost:hover:not(:disabled) {
  background: var(--bg-hover);
}

.c-btn--secondary:hover:not(:disabled) {
  border-color: var(--accent);
}

.c-btn--danger:hover:not(:disabled) {
  background: rgba(255, 107, 107, 0.12);
}

.c-btn--outline:hover:not(:disabled) {
  background: rgba(145, 71, 255, 0.15);
}

.c-btn:disabled,
.c-btn[aria-busy='true'] {
  opacity: 0.65;
  cursor: not-allowed;
}

.c-btn--full {
  width: 100%;
}

.c-btn--square {
  width: 1.15rem;
  height: 1.15rem;
  min-width: 1.15rem;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.68rem;
  line-height: 1;
}
</style>
