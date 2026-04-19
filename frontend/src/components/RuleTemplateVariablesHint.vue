<script setup lang="ts">
import type { RuleTemplateVariable } from '../api/generated/models/RuleTemplateVariable';

defineProps<{
  variables: RuleTemplateVariable[] | null;
}>();
</script>

<template>
  <span class="tpl-var-tip" tabindex="0">
    <span class="tpl-var-tip__icon" aria-label="Available template variables">?</span>
    <div class="tpl-var-tip__popover" role="tooltip">
      <p v-if="variables === null" class="tpl-var-tip__muted">Loading…</p>
      <p v-else-if="variables.length === 0" class="tpl-var-tip__muted">Could not load variables.</p>
      <ul v-else class="tpl-var-tip__list">
        <li v-for="v in variables" :key="v.name">
          <code class="tpl-var-tip__code">{{ '$' + v.name }}</code>
          <span class="tpl-var-tip__desc">{{ v.description }}</span>
        </li>
      </ul>
    </div>
  </span>
</template>

<style scoped lang="scss">
.tpl-var-tip {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-left: 0.35rem;
  vertical-align: middle;
  cursor: help;
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

.tpl-var-tip__popover {
  display: none;
  position: absolute;
  left: 0;
  top: calc(100% + 6px);
  z-index: 20;
  min-width: 14rem;
  max-width: min(22rem, 92vw);
  padding: 0.5rem 0.65rem;
  border-radius: 0.35rem;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
  font-size: 0.78rem;
  line-height: 1.35;
  color: var(--text);
  text-align: left;
}

.tpl-var-tip:hover .tpl-var-tip__popover,
.tpl-var-tip:focus-within .tpl-var-tip__popover {
  display: block;
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
  padding: 0.2rem 0;
  border-bottom: 1px solid var(--border);

  &:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }
}

.tpl-var-tip__code {
  font-family: ui-monospace, monospace;
  font-size: 0.76rem;
  color: var(--accent-bright);
  white-space: nowrap;
}

.tpl-var-tip__desc {
  color: var(--text-muted);
}

.tpl-var-tip__muted {
  margin: 0;
  color: var(--text-muted);
  font-style: italic;
}
</style>
