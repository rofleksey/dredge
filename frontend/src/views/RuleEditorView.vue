<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { DefaultService } from '../api/generated';
import { RuleActionType } from '../api/generated/models/RuleActionType';
import { RuleEventType } from '../api/generated/models/RuleEventType';
import type { RuleTemplateVariable } from '../api/generated/models/RuleTemplateVariable';
import { Button } from '../components/core';
import RuleMiddlewareRow from '../components/RuleMiddlewareRow.vue';
import RuleTemplateVariablesHint from '../components/RuleTemplateVariablesHint.vue';
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
import { notifyApiError } from '../lib/notifyApiError';
import { useTwitchAccountsStore } from '../stores/twitchAccounts';

defineOptions({ name: 'RuleEditorView' });

const twitchAccountsStore = useTwitchAccountsStore();

const route = useRoute();
const router = useRouter();

const isNew = computed(() => route.name === 'rule-new');
const ruleId = computed(() => Number.parseInt(String(route.params.id), 10));

const loading = ref(true);
const saving = ref(false);
const deleting = ref(false);

const form = ref<RuleFormState>(defaultRuleForm());
const loadedRuleId = ref<number | null>(null);
/** null = loading; empty after failed fetch */
const templateVariables = ref<RuleTemplateVariable[] | null>(null);

async function fetchTemplateVariables(): Promise<void> {
  try {
    const res = await DefaultService.listRuleTemplateVariables();
    templateVariables.value = res.variables;
  } catch {
    templateVariables.value = [];
  }
}


async function load(): Promise<void> {
  loading.value = true;
  try {
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
    notifyApiError(e, { id: 'rule-load', title: 'Rules', fallbackMessage: 'Request failed.' });
    await router.replace({ name: 'settings', query: { tab: 'rules' } });
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void load();
  void fetchTemplateVariables();
  void twitchAccountsStore.fetch();
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
  form.value.middlewares.splice(i, 1);
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
    notifyApiError(e, { id: 'rule-save', title: 'Rules', fallbackMessage: 'Request failed.' });
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
  const label = form.value.name.trim() || `#${id}`;
  if (!window.confirm(`Delete rule “${label}” (#${id})? This cannot be undone.`)) {
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
    notifyApiError(e, { id: 'rule-del', title: 'Rules', fallbackMessage: 'Request failed.' });
  } finally {
    deleting.value = false;
  }
}

function cancel(): void {
  void router.push({ name: 'settings', query: { tab: 'rules' } });
}

const pageTitle = computed(() => {
  if (isNew.value) {
    return 'New rule';
  }
  const n = form.value.name.trim();
  return n ? `Edit: ${n}` : `Edit rule #${ruleId.value}`;
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
        <label class="stack gap-setting">
          <span>Name</span>
          <input v-model="form.name" type="text" autocomplete="off" placeholder="Short label" />
        </label>
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
            type="button"
            class="btn-remove-mw"
            @click="removeMiddleware(i)"
          >
            Remove
          </button>
        </div>
        <p class="row-actions">
          <Button native-type="button" variant="secondary" @click="addMiddleware">Add middleware</Button>
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
            <span class="label-with-hint">
              <span>Message template (optional)</span>
              <RuleTemplateVariablesHint :variables="templateVariables" />
            </span>
            <textarea v-model="form.notifyText" rows="4" spellcheck="false" placeholder="Empty uses defaults for some events." />
          </label>
        </template>

        <template v-else>
          <p class="muted small">
            Message is sent to the same channel as the event. Pick which linked Twitch account sends the message, or leave
            the default (linked bot account if present, otherwise your first linked account).
          </p>
          <label class="stack gap-setting">
            <span>Send as account</span>
            <select v-model.number="form.sendAccountId">
              <option :value="0">Default (bot or first linked account)</option>
              <option v-for="a in twitchAccountsStore.accounts" :key="a.id" :value="a.id">
                @{{ a.username }} ({{ a.account_type }})
              </option>
            </select>
          </label>
          <label class="stack gap-setting">
            <span class="label-with-hint">
              <span>Message (template)</span>
              <RuleTemplateVariablesHint :variables="templateVariables" />
            </span>
            <textarea v-model="form.sendMessage" rows="3" spellcheck="false" autocomplete="off" />
          </label>
        </template>
      </section>

      <footer class="rule-editor-footer">
        <Button
          v-if="!isNew"
          native-type="button"
          variant="danger"
          :disabled="deleting"
          @click="removeRule"
        >
          Delete
        </Button>
        <div class="rule-editor-footer-actions">
          <Button native-type="button" variant="secondary" @click="cancel">Cancel</Button>
          <Button native-type="submit" :loading="saving">Save</Button>
        </div>
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

.label-with-hint {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.15rem;
  font-size: 0.82rem;
  color: var(--text-muted);
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

.small {
  font-size: 0.78rem;
}

.row-actions {
  margin: 0.35rem 0 0;
}

.rule-editor-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--border);
}

.rule-editor-footer-actions {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-left: auto;
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
