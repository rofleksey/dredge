<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { onMounted, ref, watch } from 'vue';
import { ApiError, DefaultService } from '../api/generated';
import type { AiConversation, AiMessage } from '../api/generated';
import SubmitButton from '../components/SubmitButton.vue';
import { notify } from '../lib/notify';
import { useLiveSocketStore } from '../stores/liveSocket';

const conversations = ref<AiConversation[]>([]);
const activeId = ref<number | null>(null);
const messages = ref<AiMessage[]>([]);
const draft = ref('');
const sending = ref(false);
const pending = ref<{ tool_call_id: string; tool_name: string; arguments: string } | null>(null);
const activityLog = ref<{ ts: number; text: string; kind?: string }[]>([]);

const live = useLiveSocketStore();
const { aiAgentEvents } = storeToRefs(live);

function pushActivity(text: string, kind?: string): void {
  activityLog.value.push({ ts: Date.now(), text, kind });
  if (activityLog.value.length > 120) {
    activityLog.value = activityLog.value.slice(-120);
  }
}

function describeAgentEvent(ev: Record<string, unknown>): string {
  const k = String(ev.kind ?? '');
  if (k === 'tool_attempt') {
    return `API tool: ${String(ev.tool_name ?? '')} ${String(ev.arguments ?? '').slice(0, 280)}`;
  }
  if (k === 'error') {
    return `Agent error: ${String(ev.message ?? '')}${ev.phase ? ` (${String(ev.phase)})` : ''}`;
  }
  if (k === 'llm_request') {
    return 'Calling LLM…';
  }
  if (k === 'needs_confirmation') {
    return `Needs approval: ${String(ev.tool_name ?? '')}`;
  }
  if (k === 'tool_result') {
    return `Tool finished: ${String(ev.tool_name ?? '')}`;
  }
  if (k === 'done') {
    return `Run finished: ${String(ev.reason ?? '')}`;
  }
  return k || 'event';
}

async function loadConversations(): Promise<void> {
  conversations.value = await DefaultService.listAiConversations();
  if (activeId.value === null && conversations.value.length) {
    activeId.value = conversations.value[0].id;
  }
}

async function loadMessages(): Promise<void> {
  if (activeId.value === null) {
    messages.value = [];
    return;
  }
  messages.value = await DefaultService.listAiMessages({ conversationId: activeId.value });
}

async function onNewConversation(): Promise<void> {
  try {
    const c = await DefaultService.createAiConversation({});
    await loadConversations();
    activeId.value = c.id;
    await loadMessages();
  } catch {
    notify({ id: 'ai-new', type: 'error', title: 'AI', description: 'Could not create conversation.' });
  }
}

async function onDeleteConversation(): Promise<void> {
  if (activeId.value === null) {
    return;
  }
  const id = activeId.value;
  try {
    await DefaultService.deleteAiConversation({ conversationId: id });
    pending.value = null;
    await loadConversations();
    activeId.value = conversations.value[0]?.id ?? null;
    await loadMessages();
  } catch {
    notify({ id: 'ai-del', type: 'error', title: 'AI', description: 'Could not delete conversation.' });
  }
}

async function onSend(): Promise<void> {
  const text = draft.value.trim();
  if (!text || activeId.value === null || sending.value) {
    return;
  }
  sending.value = true;
  try {
    await DefaultService.createAiMessage({
      conversationId: activeId.value,
      requestBody: { content: text },
    });
    draft.value = '';
    await loadMessages();
  } catch (e: unknown) {
    const msg =
      e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
        ? String((e.body as { message: string }).message)
        : 'Could not send message.';
    notify({ id: 'ai-send', type: 'error', title: 'AI', description: msg });
  } finally {
    sending.value = false;
  }
}

async function onStop(): Promise<void> {
  if (activeId.value === null) {
    return;
  }
  try {
    await DefaultService.stopAiAgent({ conversationId: activeId.value });
    pushActivity('Stop requested', 'stop');
  } catch {
    notify({ id: 'ai-stop', type: 'error', title: 'AI', description: 'Could not stop agent.' });
  }
}

async function onConfirm(approve: boolean): Promise<void> {
  if (!pending.value || activeId.value === null) {
    return;
  }
  try {
    await DefaultService.confirmAiTool({
      conversationId: activeId.value,
      requestBody: { tool_call_id: pending.value.tool_call_id, approve },
    });
    pending.value = null;
    await loadMessages();
  } catch (e: unknown) {
    const msg =
      e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
        ? String((e.body as { message: string }).message)
        : 'Could not confirm tool.';
    notify({ id: 'ai-confirm', type: 'error', title: 'AI', description: msg });
  }
}

watch(activeId, () => {
  pending.value = null;
  void loadMessages();
});

watch(
  aiAgentEvents,
  (evs) => {
    const last = evs[evs.length - 1];
    if (!last || last.type !== 'ai_agent') {
      return;
    }
    const cid =
      typeof last.conversation_id === 'number'
        ? last.conversation_id
        : Number(last.conversation_id);
    if (activeId.value !== null && Number.isFinite(cid) && cid !== activeId.value) {
      return;
    }
    const kind = String(last.kind ?? '');
    if (kind === 'needs_confirmation') {
      pending.value = {
        tool_call_id: String(last.tool_call_id ?? ''),
        tool_name: String(last.tool_name ?? ''),
        arguments: String(last.arguments ?? ''),
      };
    }
    if (
      kind === 'tool_result' ||
      kind === 'done' ||
      kind === 'message' ||
      kind === 'error' ||
      kind === 'needs_confirmation'
    ) {
      void loadMessages();
    }
    if (
      kind === 'tool_attempt' ||
      kind === 'llm_request' ||
      kind === 'error' ||
      kind === 'needs_confirmation' ||
      kind === 'done'
    ) {
      pushActivity(describeAgentEvent(last as Record<string, unknown>), kind);
    }
  },
  { deep: true },
);

onMounted(async () => {
  try {
    await loadConversations();
    await loadMessages();
  } catch {
    notify({ id: 'ai-load', type: 'error', title: 'AI', description: 'Failed to load (admin only?).' });
  }
});

function convLabel(c: AiConversation): string {
  if (c.title) {
    return c.title;
  }
  return `Conversation #${c.id}`;
}

function formatMeta(m: Record<string, unknown>): string {
  if (!m || typeof m !== 'object') {
    return '';
  }
  if (m.error) {
    return JSON.stringify(m);
  }
  if (m.tool_calls) {
    return '[tool_calls]';
  }
  if (m.system) {
    return '[system]';
  }
  return '';
}
</script>

<template>
  <div class="ai-wrap">
    <div class="ai-layout">
      <aside class="ai-sidebar">
        <div class="ai-sidebar-head">
          <h2>Chats</h2>
          <button type="button" class="btn-ghost" @click="onNewConversation">New</button>
        </div>
        <ul class="ai-conv-list">
          <li v-for="c in conversations" :key="c.id">
            <button
              type="button"
              class="ai-conv-btn"
              :class="{ active: activeId === c.id }"
              @click="activeId = c.id"
            >
              {{ convLabel(c) }}
            </button>
          </li>
        </ul>
        <button
          type="button"
          class="btn-danger-ghost"
          :disabled="activeId === null"
          @click="onDeleteConversation"
        >
          Delete chat
        </button>
      </aside>
      <main class="ai-main">
        <header class="ai-toolbar">
          <h1>AI</h1>
          <div class="ai-toolbar-actions">
            <button type="button" class="btn-ghost" @click="onStop">Stop agent</button>
          </div>
        </header>

        <div v-if="pending" class="ai-pending panel">
          <p>
            <strong>{{ pending.tool_name }}</strong> requires confirmation.
          </p>
          <pre class="ai-args">{{ pending.arguments }}</pre>
          <div class="row">
            <button type="button" class="btn-primary" @click="onConfirm(true)">Approve</button>
            <button type="button" class="btn-ghost" @click="onConfirm(false)">Reject</button>
          </div>
        </div>

        <div class="ai-messages panel">
          <div v-for="m in messages" :key="m.id" class="ai-msg" :data-role="m.role">
            <span class="ai-role">{{ m.role }}</span>
            <pre class="ai-content">{{ m.content }}</pre>
            <p v-if="formatMeta(m.metadata)" class="ai-meta muted">{{ formatMeta(m.metadata) }}</p>
          </div>
          <p v-if="!messages.length" class="muted">No messages yet. Configure the model in Settings → AI, then send a prompt.</p>
        </div>

        <div class="ai-activity panel">
          <h3>Agent activity</h3>
          <ul class="ai-log">
            <li v-for="(e, i) in activityLog" :key="i">
              <span class="muted">{{ new Date(e.ts).toLocaleTimeString() }}</span>
              {{ e.text }}
            </li>
          </ul>
        </div>

        <form class="ai-compose panel" @submit.prevent="onSend">
          <textarea v-model="draft" rows="3" placeholder="Message the agent…" />
          <SubmitButton :loading="sending" :disabled="activeId === null">Send</SubmitButton>
        </form>
      </main>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ai-wrap {
  padding: 0.75rem 1rem;
  max-width: 1200px;
  margin: 0 auto;
}

.ai-layout {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 1rem;
  min-height: 60vh;
}

.ai-sidebar {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.75rem;
  background: var(--bg-elevated);
}

.ai-sidebar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  h2 {
    margin: 0;
    font-size: 1rem;
  }
}

.ai-conv-list {
  list-style: none;
  margin: 0;
  padding: 0;
  flex: 1;
  overflow: auto;
  max-height: 50vh;
}

.ai-conv-btn {
  width: 100%;
  text-align: left;
  padding: 0.35rem 0.5rem;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font: inherit;
  &.active {
    border-color: var(--border);
    background: var(--bg-base);
  }
}

.btn-danger-ghost {
  border: 1px solid var(--danger, #c44);
  color: var(--danger, #c44);
  background: transparent;
  border-radius: 6px;
  padding: 0.35rem 0.5rem;
  cursor: pointer;
  font: inherit;
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.ai-main {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ai-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  h1 {
    margin: 0;
    font-size: 1.25rem;
  }
}

.panel {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.75rem;
  background: var(--bg-elevated);
}

.ai-messages {
  flex: 1;
  overflow: auto;
  max-height: 42vh;
}

.ai-msg {
  margin-bottom: 0.75rem;
  &[data-role='user'] .ai-role {
    color: var(--accent);
  }
  &[data-role='assistant'] .ai-role {
    color: var(--text);
  }
  &[data-role='tool'] .ai-role {
    color: var(--muted, #888);
  }
}

.ai-role {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.ai-content {
  margin: 0.25rem 0 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
  font-size: 0.95rem;
}

.ai-meta {
  font-size: 0.8rem;
  margin: 0.25rem 0 0;
}

.ai-pending {
  border-color: var(--accent);
}

.ai-args {
  font-size: 0.8rem;
  overflow: auto;
  max-height: 8rem;
  margin: 0.5rem 0;
}

.ai-activity {
  max-height: 14rem;
  overflow: auto;
  h3 {
    margin: 0 0 0.5rem;
    font-size: 0.95rem;
  }
}

.ai-log {
  list-style: none;
  margin: 0;
  padding: 0;
  font-size: 0.85rem;
  li {
    margin-bottom: 0.35rem;
  }
}

.ai-compose {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  textarea {
    width: 100%;
    resize: vertical;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    padding: 0.5rem;
    font: inherit;
  }
}

.muted {
  opacity: 0.75;
}

.row {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

@media (max-width: 800px) {
  .ai-layout {
    grid-template-columns: 1fr;
  }
}

.btn-primary {
  border: none;
  border-radius: 6px;
  padding: 0.4rem 0.75rem;
  background: var(--accent);
  color: var(--bg-base, #0d0d0f);
  cursor: pointer;
  font: inherit;
}
</style>
