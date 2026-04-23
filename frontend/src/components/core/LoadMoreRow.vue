<script setup lang="ts">
import { computed } from 'vue';

type LoadMoreVariant = 'solid' | 'ghost';
type LoadMoreRowSpacing = 'default' | 'tight' | 'relaxed';

const props = withDefaults(
  defineProps<{
    loading?: boolean;
    disabled?: boolean;
    variant?: LoadMoreVariant;
    rowSpacing?: LoadMoreRowSpacing;
    /** When false and `exhaustedText` is set, button shows exhausted label and stays disabled. */
    hasMore?: boolean;
    idleText?: string;
    loadingText?: string;
    exhaustedText?: string;
  }>(),
  {
    loading: false,
    disabled: false,
    variant: 'solid',
    rowSpacing: 'default',
    hasMore: true,
    idleText: 'Load more',
    loadingText: 'Loading…',
    exhaustedText: '',
  },
);

const emit = defineEmits<{ click: [] }>();

const isExhausted = computed(
  () => Boolean(props.exhaustedText) && !props.hasMore && !props.loading,
);

const label = computed(() => {
  if (props.loading) {
    return props.loadingText;
  }
  if (isExhausted.value) {
    return props.exhaustedText;
  }
  return props.idleText;
});

const buttonDisabled = computed(
  () => props.loading || props.disabled || isExhausted.value,
);

const rowClass = computed(() => {
  const list = ['load-more-row'];
  if (props.rowSpacing === 'tight') {
    list.push('load-more-row--tight');
  } else if (props.rowSpacing === 'relaxed') {
    list.push('load-more-row--relaxed');
  }
  return list;
});

const btnClass = computed(() => [
  'load-more-btn',
  props.variant === 'ghost' ? 'load-more-btn--ghost' : 'load-more-btn--solid',
]);

function onClick(): void {
  if (buttonDisabled.value) {
    return;
  }
  emit('click');
}
</script>

<template>
  <div :class="rowClass">
    <button type="button" :class="btnClass" :disabled="buttonDisabled" @click="onClick">
      {{ label }}
    </button>
  </div>
</template>

<style scoped lang="scss">
.load-more-row {
  margin-top: 0.5rem;

  &--tight {
    margin-top: 0.35rem;
  }

  &--relaxed {
    margin-top: 0.65rem;
  }
}

.load-more-btn {
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.85rem;

  &:disabled {
    cursor: not-allowed;
  }

  &--solid {
    padding: 0.4rem 0.85rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);

    &:disabled {
      opacity: 0.5;
    }
  }

  &--ghost {
    padding: 0.35rem 0.75rem;
    border: 1px dashed var(--border);
    background: transparent;
    color: var(--text-muted);

    &:hover:not(:disabled) {
      border-color: var(--accent);
      color: var(--accent);
    }

    &:disabled {
      opacity: 0.5;
    }
  }
}
</style>
