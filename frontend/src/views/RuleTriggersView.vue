<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ApiError, DefaultService } from '../api/generated';
import type { RuleTrigger } from '../api/generated';
import { notify } from '../lib/notify';

defineOptions({ name: 'RuleTriggersView' });

const pageSize = 50;
const loading = ref(false);
const loadingMore = ref(false);
const hasMore = ref(true);
const entries = ref<RuleTrigger[]>([]);

const totalLoadedLabel = computed(() => `${entries.value.length.toLocaleString()} loaded`);

function createdAtLabel(ts: string): string {
  const d = new Date(ts);
  if (!Number.isFinite(d.getTime())) {
    return ts;
  }
  return d.toLocaleString();
}

function ruleIdLabel(e: RuleTrigger): string {
  if (e.rule_id != null && e.rule_id !== undefined) {
    return `#${e.rule_id}`;
  }
  return '—';
}

async function fetchFirst(): Promise<void> {
  loading.value = true;
  try {
    const list = await DefaultService.listRuleTriggers({ limit: pageSize });
    entries.value = list;
    hasMore.value = list.length === pageSize;
  } catch (e) {
    entries.value = [];
    hasMore.value = false;
    const msg =
      e instanceof ApiError && e.body && typeof e.body.message === 'string'
        ? e.body.message
        : 'Could not load rule triggers.';
    notify({
      id: 'rule-triggers-load',
      type: 'error',
      title: 'Rule triggers',
      description: msg,
    });
  } finally {
    loading.value = false;
  }
}

async function fetchMore(): Promise<void> {
  if (!hasMore.value || loadingMore.value || entries.value.length === 0) {
    return;
  }

  loadingMore.value = true;
  try {
    const last = entries.value[entries.value.length - 1];
    const next = await DefaultService.listRuleTriggers({
      limit: pageSize,
      cursorCreatedAt: last.created_at,
      cursorId: last.id,
    });
    const seen = new Set(entries.value.map((x) => x.id));
    for (const row of next) {
      if (!seen.has(row.id)) {
        entries.value.push(row);
      }
    }
    hasMore.value = next.length === pageSize;
  } catch {
    notify({
      id: 'rule-triggers-more',
      type: 'error',
      title: 'Rule triggers',
      description: 'Could not load more.',
    });
  } finally {
    loadingMore.value = false;
  }
}

onMounted(() => {
  void fetchFirst();
});
</script>

<template>
  <div class="page">
    <header class="head">
      <h1>Rule triggers</h1>
      <span class="muted small">{{ totalLoadedLabel }}</span>
    </header>

    <p v-if="loading" class="muted">Loading...</p>
    <p v-else-if="entries.length === 0" class="muted">No rule triggers yet.</p>

    <ul v-else class="list">
      <li v-for="e in entries" :key="e.id" class="item">
        <div class="item-top">
          <span class="tag">{{ e.action_type }}</span>
          <span class="tag tag-muted">{{ e.trigger_event }}</span>
        </div>
        <p class="meta">
          #{{ e.id }} · {{ createdAtLabel(e.created_at) }} · rule {{ ruleIdLabel(e) }}
          <template v-if="e.rule_name"> · {{ e.rule_name }}</template>
        </p>
        <p class="body">{{ e.display_text }}</p>
      </li>
    </ul>

    <div v-if="!loading && entries.length > 0" class="more-row">
      <button type="button" class="btn-more" :disabled="loadingMore || !hasMore" @click="fetchMore">
        {{ loadingMore ? 'Loading...' : hasMore ? 'Show more' : 'No more' }}
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

h1 {
  margin: 0;
  font-size: 1.15rem;
}

.list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.item {
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  background: var(--bg-elevated);
  padding: 0.55rem 0.6rem;
}

.item-top {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.tag {
  font-size: 0.76rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--accent-bright);
}

.tag-muted {
  color: var(--text-muted);
  text-transform: none;
  letter-spacing: 0;
}

.meta {
  margin: 0.25rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.body {
  margin: 0.35rem 0 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.82rem;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-word;
}

.more-row {
  margin-top: 0.65rem;
}

.btn-more {
  padding: 0.42rem 0.85rem;
  border: 1px solid var(--border);
  border-radius: 0.25rem;
  background: var(--bg-base);
  color: var(--text);
  cursor: pointer;

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
}

.muted {
  color: var(--text-muted);
}

.small {
  font-size: 0.78rem;
}
</style>
