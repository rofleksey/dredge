<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { ApiError, DefaultService } from '../api/generated';
import { RuleActionType } from '../api/generated/models/RuleActionType';
import { RuleEventType } from '../api/generated/models/RuleEventType';
import RuleMiddlewareRow from '../components/RuleMiddlewareRow.vue';
import SubmitButton from '../components/SubmitButton.vue';
import {
  defaultMiddlewareRow,
  defaultRuleForm,
  formStateToCreateRequest,
  formStateToUpdateRequest,
  ruleToFormState,
  type RuleFormState,
  validateRuleForm,
} from '../lib/ruleForm';
import { notify } from '../lib/notify';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';

defineOptions({ name: 'RuleEditorView' });

const route = useRoute();
const router = useRouter();
const twitchStore = useTwitchAccountsStore();

const isNew = computed(() => route.name === 'rule-new');
const ruleId = computed(() => Number.parseInt(String(route.params.id), 10));

const loading = ref(true);
const saving = ref(false);
const deleting = ref(false);

const form = ref<RuleFormState>(defaultRuleForm());
const loadedRuleId = ref<number | null>(null);

function notifyErr(e: unknown, id: string, title: string): void {
  const msg =
    e instanceof ApiError && e.body && typeof (e.body as { message?: string }).message === 'string'
      ? (e.body as { message: string }).message
      : 'Request failed.';
  notify({ id, type: 'error', title, description: msg });
}

async function load(): Promise<void> {
  loading.value = true;
  try {
    await twitchStore.fetch();
    if (isNew.value) {
      form.value = defaultRuleForm();
      loadedRuleId.value = null;
      return;
    }
    if (!Number.isFinite(ruleId.value)) {
      notify({ id: 'rule-load', type: 'error', title: 'Rules', description: 'Invalid rule id.' });
      await router.replace({ name: 'settings', query: { tab: 'rules' } });
      return;
    }
    const list = await DefaultService.listRules();
    const found = list.find((r) => r.id === ruleId.value);
    if (!found) {
      notify({ id: 'rule-load', type: 'error', title: 'Rules', description: 'Rule not found.' });
      await router.replace({ name: 'settings', query: { tab: 'rules' } });
      return;
    }
    form.value = ruleToFormState(found);
    loadedRuleId.value = found.id;
  } catch (e) {
    notifyErr(e, 'rule-load', 'Rules');
    await router.replace({ name: 'settings', query: { tab: 'rules' } });
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
});

watch(
  () => route.fullPath,
  () => {
    void load();
  },
);

function addMiddleware(): void {
  form.value.middlewares.push(defaultMiddlewareRow('match_regex'));
}

function removeMiddleware(i: number): void {
  if (form.value.middlewares.length <= 1) {
    return;
  }
  form.value.middlewares.splice(i, 1);
}

function onPickAccount(ev: Event): void {
  const v = (ev.target as HTMLSelectElement).value;
  if (v) {
    form.value.sendAccountId = v;
  }
}

async function save(): Promise<void> {
  const err = validateRuleForm(form.value);
  if (err) {
    notify({ id: 'rule-save', type: 'error', title: 'Rules', description: err });
    return;
  }
  if (saving.value) {
    return;
  }
  saving.value = true;
  try {
    if (isNew.value) {
      await DefaultService.createRule({
        requestBody: formStateToCreateRequest(form.value),
      });
      notify({ id: 'rule-save', type: 'success', title: 'Rules', description: 'Rule created.' });
    } else {
      const id = loadedRuleId.value ?? ruleId.value;
      await DefaultService.updateRule({
        requestBody: formStateToUpdateRequest(id, form.value),
      });
      notify({ id: 'rule-save', type: 'success', title: 'Rules', description: 'Rule saved.' });
    }
    await router.push({ name: 'settings', query: { tab: 'rules' } });
  } catch (e) {
    notifyErr(e, 'rule-save', 'Rules');
  } finally {
    saving.value = false;
  }
}

async function removeRule(): Promise<void> {
  if (isNew.value) {
    return;
  }
  const id = loadedRuleId.value ?? ruleId.value;
  if (!Number.isFinite(id)) {
    return;
  }
  if (!window.confirm(`Delete rule #${id}? This cannot be undone.`)) {
    return;
  }
  if (deleting.value) {
    return;
  }
  deleting.value = true;
  try {
    await DefaultService.deleteRule({ requestBody: { id } });
    notify({ id: 'rule-del', type: 'success', title: 'Rules', description: 'Rule deleted.' });
    await router.push({ name: 'settings', query: { tab: 'rules' } });
  } catch (e) {
    notifyErr(e, 'rule-del', 'Rules');
  } finally {
    deleting.value = false;
  }
}

function cancel(): void {
  void router.push({ name: 'settings', query: { tab: 'rules' } });
}

const pageTitle = computed(() => (isNew.value ? 'New rule' : `Edit rule #${ruleId.value}`));

const accountPickValue = computed(() => {
  const id = form.value.sendAccountId.trim();
  if (!id) {
    return '';
  }
  const n = twitchStore.accounts.some((a) => String(a.id) === id);
  return n ? id : '';
});
</script>

<template>
  <div class="rule-editor-wrap">
    <div v-if="loading" class="rule-editor muted">Loading…</div>
    <form v-else class="rule-editor" @submit.prevent="save">
      <p class="rule-editor-back">
        <RouterLink class="back-link" :to="{ name: 'settings', query: { tab: 'rules' } }">← Settings · Rules</RouterLink>
      </p>
      <h1>{{ pageTitle }}</h1>
      <p class="hint">
        Events: <code>chat_message</code> (IRC), <code>stream_start</code> / <code>stream_end</code> (Helix), or
        <code>interval</code> (periodic tick; set channel and seconds). Middlewares run in order; cooldown is applied
        after other filters match.
      </p>

      <section class="panel">
        <h2>General</h2>
        <label class="row-inline">
          <input v-model="form.enabled" type="checkbox" />
          <span>Enabled</span>
        </label>
        <label class="row-inline">
          <input v-model="form.useSharedPool" type="checkbox" />
          <span>Use shared pool</span>
        </label>
      </section>

      <section class="panel">
        <h2>Event</h2>
        <label class="stack gap-setting">
          <span>Event type</span>
          <select v-model="form.eventType">
            <option :value="RuleEventType.CHAT_MESSAGE">chat_message</option>
            <option :value="RuleEventType.STREAM_START">stream_start</option>
            <option :value="RuleEventType.STREAM_END">stream_end</option>
            <option :value="RuleEventType.INTERVAL">interval</option>
          </select>
        </label>
        <template v-if="form.eventType === RuleEventType.INTERVAL">
          <label class="stack gap-setting">
            <span>Interval (seconds)</span>
            <input v-model="form.intervalSeconds" type="text" inputmode="numeric" autocomplete="off" />
          </label>
          <label class="stack gap-setting">
            <span>Channel login</span>
            <input v-model="form.intervalChannel" type="text" autocomplete="off" placeholder="monitored channel" />
          </label>
        </template>
      </section>

      <section class="panel">
        <h2>Middlewares</h2>
        <p class="muted small">Order matters. Add filters, a match condition, then optionally cooldown.</p>
        <div v-for="(mw, i) in form.middlewares" :key="mw.key" class="mw-line">
          <RuleMiddlewareRow v-model="form.middlewares[i]!" />
          <button
            v-if="form.middlewares.length > 1"
            type="button"
            class="btn-remove-mw"
            @click="removeMiddleware(i)"
          >
            Remove
          </button>
        </div>
        <p class="row-actions">
          <button type="button" class="btn-secondary" @click="addMiddleware">Add middleware</button>
        </p>
      </section>

      <section class="panel">
        <h2>Action</h2>
        <label class="stack gap-setting">
          <span>Action type</span>
          <select v-model="form.actionType">
            <option :value="RuleActionType.NOTIFY">notify</option>
            <option :value="RuleActionType.SEND_CHAT">send_chat</option>
          </select>
        </label>

        <template v-if="form.actionType === RuleActionType.NOTIFY">
          <label class="stack gap-setting">
            <span>Message template (optional; uses <code>$CHANNEL</code> <code>$USERNAME</code> <code>$TEXT</code> <code>$TITLE</code> <code>$RULE_ID</code>)</span>
            <textarea v-model="form.notifyText" rows="4" spellcheck="false" placeholder="Empty uses defaults for some events." />
          </label>
        </template>

        <template v-else>
          <label class="stack gap-setting">
            <span>Linked Twitch account</span>
            <select :value="accountPickValue" @change="onPickAccount">
              <option value="">— Pick account (Settings → Twitch) —</option>
              <option v-for="a in twitchStore.accounts" :key="a.id" :value="String(a.id)">
                {{ a.username }} ({{ a.account_type }}) · id {{ a.id }}
              </option>
            </select>
          </label>
          <label class="stack gap-setting">
            <span>Account id</span>
            <input v-model="form.sendAccountId" type="number" min="1" step="1" autocomplete="off" />
          </label>
          <label class="stack gap-setting">
            <span>Channel (template)</span>
            <input v-model="form.sendChannel" type="text" autocomplete="off" placeholder="$CHANNEL" />
          </label>
          <label class="stack gap-setting">
            <span>Message (template)</span>
            <textarea v-model="form.sendMessage" rows="3" spellcheck="false" autocomplete="off" />
          </label>
        </template>
      </section>

      <footer class="rule-editor-footer">
        <button type="button" class="btn-secondary" @click="cancel">Cancel</button>
        <SubmitButton :loading="saving">Save</SubmitButton>
        <button
          v-if="!isNew"
          type="button"
          class="btn-danger"
          :disabled="deleting"
          @click="removeRule"
        >
          Delete
        </button>
      </footer>
    </form>
  </div>
</template>

<style scoped lang="scss">
.rule-editor-wrap {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 1rem 1rem 2rem;
  min-height: 0;
  overflow-y: auto;
}

.rule-editor {
  width: 100%;
  max-width: 640px;
}

.rule-editor-back {
  margin: 0 0 0.5rem;
}

.back-link {
  color: var(--accent-bright);
  font-size: 0.88rem;
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }
}

h1 {
  font-size: 1.25rem;
  margin: 0 0 0.5rem;
  text-align: center;
  color: var(--accent-bright);
}

.hint {
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.45;
  margin: 0 0 1rem;
}

.panel {
  border-bottom: 1px solid var(--border);
  padding-bottom: 1rem;
  margin-bottom: 0.75rem;
}

h2 {
  font-size: 1rem;
  margin: 0 0 0.5rem;
  color: var(--accent-bright);
}

.stack.gap-setting {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-bottom: 0.65rem;

  span {
    font-size: 0.82rem;
    color: var(--text-muted);
  }

  input,
  select,
  textarea {
    width: 100%;
    padding: 0.4rem 0.55rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    color: var(--text);
    font-size: 0.88rem;
  }

  textarea {
    font-family: ui-monospace, monospace;
    font-size: 0.82rem;
    resize: vertical;
    min-height: 5rem;
  }
}

.row-inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  margin-bottom: 0.45rem;
}

.muted {
  color: var(--text-muted);
}

.small {
  font-size: 0.78rem;
}

.row-actions {
  margin: 0.35rem 0 0;
}

.btn-secondary {
  padding: 0.45rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  color: var(--text);
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  margin-right: 0.5rem;

  &:hover {
    background: var(--bg-hover);
    border-color: var(--accent);
  }
}

.btn-danger {
  padding: 0.45rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid #c0392b;
  background: rgba(192, 57, 43, 0.2);
  color: #ffb4a8;
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;

  &:hover:not(:disabled) {
    background: rgba(192, 57, 43, 0.35);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.rule-editor-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
  padding-top: 0.5rem;
}

.mw-line {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.5rem;
  margin-bottom: 0.35rem;
}

.mw-line :deep(.mw-row) {
  flex: 1 1 220px;
  min-width: 0;
}

.btn-remove-mw {
  flex: 0 0 auto;
  margin-top: 1.85rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: var(--bg-base);
  color: var(--text-muted);
  cursor: pointer;

  &:hover {
    color: var(--text);
    border-color: #c0392b;
  }
}
</style>
