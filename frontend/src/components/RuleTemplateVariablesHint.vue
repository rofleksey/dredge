<script setup lang="ts">
import { ref } from 'vue';
import type { RuleTemplateVariable } from '../api/generated/models/RuleTemplateVariable';
import AppModal from './AppModal.vue';

defineProps<{
  variables: RuleTemplateVariable[] | null;
}>();

const modalOpen = ref(false);
</script>

<template>
  <span class="tpl-var-tip">
    <button
      type="button"
      class="tpl-var-tip__btn"
      aria-haspopup="dialog"
      :aria-expanded="modalOpen"
      aria-label="Template variables — open list"
      @click="modalOpen = true"
    >
      <span class="tpl-var-tip__icon" aria-hidden="true">?</span>
    </button>

    <AppModal title="Template variables" :open="modalOpen" wide @close="modalOpen = false">
      <p v-if="variables === null" class="tpl-var-tip__muted">Loading…</p>
      <p v-else-if="variables.length === 0" class="tpl-var-tip__muted">Could not load variables.</p>
      <ul v-else class="tpl-var-tip__list">
        <li v-for="v in variables" :key="v.name">
          <code class="tpl-var-tip__code">{{ '$' + v.name }}</code>
          <span class="tpl-var-tip__desc">{{ v.description }}</span>
        </li>
      </ul>
    </AppModal>
  </span>
</template>

<style scoped lang="scss">
.tpl-var-tip {
  display: inline-flex;
  align-items: center;
  margin-left: 0.35rem;
  vertical-align: middle;
}

.tpl-var-tip__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 50%;
  outline: none;

  &:focus-visible .tpl-var-tip__icon {
    box-shadow: 0 0 0 2px var(--accent, #6cf);
  }
}

.tpl-var-tip__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.1rem;
  height: 1.1rem;
  border-radius: 50%;
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1;
  color: var(--text-muted);
  border: 1px solid var(--border);
  background: var(--bg-base);
  user-select: none;
}

.tpl-var-tip__btn:hover .tpl-var-tip__icon {
  background: var(--bg-hover);
  color: var(--text);
}

.tpl-var-tip__list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.tpl-var-tip__list li {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.35rem 0.5rem;
  align-items: start;
  padding: 0.35rem 0;
  border-bottom: 1px solid var(--border);

  &:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }
}

.tpl-var-tip__code {
  font-family: ui-monospace, monospace;
  font-size: 0.82rem;
  color: var(--accent-bright);
  white-space: nowrap;
}

.tpl-var-tip__desc {
  color: var(--text-muted);
  font-size: 0.88rem;
  line-height: 1.4;
}

.tpl-var-tip__muted {
  margin: 0;
  color: var(--text-muted);
  font-style: italic;
}
</style>
