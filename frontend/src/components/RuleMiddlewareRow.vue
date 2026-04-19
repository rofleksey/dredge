<script setup lang="ts">
import { MIDDLEWARE_TYPES, type MiddlewareFormRow } from '../lib/ruleForm';

const row = defineModel<MiddlewareFormRow>({ required: true });

function onTypeChange(ev: Event): void {
  const v = (ev.target as HTMLSelectElement).value;
  if (!MIDDLEWARE_TYPES.includes(v as (typeof MIDDLEWARE_TYPES)[number])) {
    return;
  }
  row.value = {
    ...row.value,
    type: v as MiddlewareFormRow['type'],
  };
}
</script>

<template>
  <div class="mw-row">
    <div class="mw-row-head">
      <label class="mw-type">
        <span class="mw-label">Type</span>
        <select :value="row.type" @change="onTypeChange">
          <option value="filter_channel">filter_channel</option>
          <option value="filter_user">filter_user</option>
          <option value="match_regex">match_regex</option>
          <option value="contains_word">contains_word</option>
          <option value="cooldown">cooldown</option>
        </select>
      </label>
    </div>

    <template v-if="row.type === 'filter_channel' || row.type === 'filter_user'">
      <label class="stack tight">
        <span>Include logins (space or comma separated)</span>
        <input v-model="row.includeLogins" type="text" autocomplete="off" placeholder="e.g. channel1 channel2" />
      </label>
      <label class="stack tight">
        <span>Exclude logins</span>
        <input v-model="row.excludeLogins" type="text" autocomplete="off" />
      </label>
      <label v-if="row.type === 'filter_channel'" class="row-inline">
        <input v-model="row.requireOnline" type="checkbox" />
        <span>Require channel live (Helix)</span>
      </label>
    </template>

    <template v-else-if="row.type === 'match_regex'">
      <label class="stack tight">
        <span>Pattern</span>
        <input v-model="row.pattern" type="text" autocomplete="off" placeholder="regular expression" />
      </label>
      <label class="row-inline">
        <input v-model="row.caseInsensitive" type="checkbox" />
        <span>Case insensitive</span>
      </label>
    </template>

    <template v-else-if="row.type === 'contains_word'">
      <label class="stack tight">
        <span>Words (one per line or comma-separated)</span>
        <textarea v-model="row.words" rows="3" spellcheck="false" autocomplete="off" />
      </label>
    </template>

    <template v-else-if="row.type === 'cooldown'">
      <label class="stack tight">
        <span>Seconds</span>
        <input v-model="row.seconds" type="text" inputmode="numeric" autocomplete="off" placeholder="e.g. 300" />
      </label>
    </template>
  </div>
</template>

<style scoped lang="scss">
.mw-row {
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  padding: 0.75rem;
  margin-bottom: 0.65rem;
  background: var(--bg-elevated);
}

.mw-row-head {
  margin-bottom: 0.5rem;
}

.mw-type {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.mw-type select {
  max-width: 100%;
}

.mw-label {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.stack.tight {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.5rem;

  span {
    font-size: 0.82rem;
    color: var(--text-muted);
  }

  input,
  textarea {
    width: 100%;
    padding: 0.35rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.88rem;
  }

  textarea {
    font-family: ui-monospace, monospace;
    font-size: 0.82rem;
    resize: vertical;
    min-height: 4rem;
  }
}

.row-inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  margin-bottom: 0.35rem;
}
</style>
