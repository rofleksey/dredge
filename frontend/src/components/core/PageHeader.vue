<script setup lang="ts">
import { computed, useSlots } from 'vue';

type PageHeaderSize = 'compact' | 'large';
type PageHeaderLayout = 'inline' | 'stacked';

const props = withDefaults(
  defineProps<{
    title?: string;
    subtitle?: string;
    totalCount?: number | null;
    size?: PageHeaderSize;
    layout?: PageHeaderLayout;
  }>(),
  {
    title: '',
    subtitle: '',
    totalCount: undefined,
    size: 'compact',
    layout: 'inline',
  },
);

const slots = useSlots();

const useBodySlot = computed(() => Boolean(slots.body));

const headerClass = computed(() => {
  const list = ['page-header'];
  if (useBodySlot.value) {
    list.push('page-header--body');
    list.push(props.layout === 'stacked' ? 'page-header--stacked' : 'page-header--inline');
    return list;
  }
  list.push(props.layout === 'stacked' ? 'page-header--stacked' : 'page-header--inline');
  list.push(props.size === 'large' ? 'page-header--size-large' : 'page-header--size-compact');
  return list;
});

const showCountPill = computed(() => props.totalCount != null);
</script>

<template>
  <header :class="headerClass">
    <template v-if="useBodySlot">
      <slot name="body" />
    </template>
    <template v-else-if="layout === 'stacked'">
      <h1 class="page-header__title">
        <slot name="title">{{ title }}</slot>
      </h1>
      <p v-if="subtitle" class="page-header__subtitle muted">{{ subtitle }}</p>
      <slot name="below" />
    </template>
    <template v-else>
      <h1 class="page-header__title">
        <slot name="title">
          {{ title }}
          <span v-if="showCountPill" class="page-header__count-pill">
            {{ (totalCount as number).toLocaleString() }} total
          </span>
        </slot>
      </h1>
      <div v-if="$slots.trailing" class="page-header__trailing">
        <slot name="trailing" />
      </div>
    </template>
  </header>
</template>

<style scoped lang="scss">
.page-header {
  margin-bottom: 0.75rem;

  &--inline {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.75rem;
  }

  &--stacked.page-header--size-large {
    .page-header__title {
      margin: 0 0 0.25rem;
      font-size: 1.35rem;
      font-weight: 600;
    }
  }

  &--stacked:not(.page-header--size-large) {
    .page-header__title {
      margin: 0 0 0.25rem;
    }
  }

  &--size-compact .page-header__title {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 600;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.5rem;
  }

  &--body.page-header--stacked {
    margin-bottom: 0.75rem;
  }

  &--body.page-header--inline {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  &__subtitle {
    margin: 0;
    font-size: inherit;
  }

  &__trailing {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
  }

  &__count-pill {
    font-size: 0.78rem;
    font-weight: 500;
    color: var(--text-muted);
  }
}
</style>
