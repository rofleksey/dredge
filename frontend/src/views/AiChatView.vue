<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { AiMessage, DefaultService } from '../api/generated';
import type { AiConversation } from '../api/generated';
import ChatMessageLine from '../components/ChatMessageLine.vue';
import { Button } from '../components/core';
import { notify } from '../lib/notify';
import { notifyApiError } from '../lib/notifyApiError';
import { useLiveSocketStore } from '../stores/liveSocket';

type ActivityEntry = { ts: number; text: string; kind?: string };

type TimelineRow =
  | { type: 'message'; message: AiMessage }
  | { type: 'activity'; entry: ActivityEntry };

const conversations = ref<AiConversation[]>([]);
const activeId = ref<number | null>(null);
const messages = ref<AiMessage[]>([]);
const draft = ref('');
const sending = ref(false);
/** Agent run in progress (after send / resume until done, error, or approval prompt). */
const agentProcessing = ref(false);
const pending = ref<{ tool_call_id: string; tool_name: string; arguments: string } | null>(null);
const activityLog = ref<ActivityEntry[]>([]);
const chatEl = ref<HTMLElement | null>(null);

const live = useLiveSocketStore();
const { aiAgentEvents } = storeToRefs(live);

function pushActivity(text: string, kind?: string): void {
  activityLog.value.push({ ts: Date.now(), text, kind });
  if (activityLog.value.length > 120) {
    activityLog.value = activityLog.value.slice(-120);
  }
}

const showAgentSpinner = computed(() => sending.value || agentProcessing.value);

function describeAgentEvent(ev: Record<string, unknown>): string {
  const k = String(ev.kind ?? '');
  if (k === 'tool_attempt') {
    return `API tool: ${String(ev.tool_name ?? '')}`;
  }
  if (k === 'error') {
    return `Agent error: ${String(ev.message ?? '')}${ev.phase ? ` (${String(ev.phase)})` : ''}`;
  }
  if (k === 'needs_confirmation') {
    return `Needs approval: ${String(ev.tool_name ?? '')}`;
  }
  if (k === 'tool_result') {
    return `Tool finished: ${String(ev.tool_name ?? '')}`;
  }
  return k || 'event';
}

const timelineRows = computed<TimelineRow[]>(() => {
  const rows: TimelineRow[] = [];
  for (const m of messages.value) {
    rows.push({ type: 'message', message: m });
  }
  for (const e of activityLog.value) {
    rows.push({ type: 'activity', entry: e });
  }
  rows.sort((a, b) => {
    const ta = a.type === 'message' ? new Date(a.message.created_at).getTime() : a.entry.ts;
    const tb = b.type === 'message' ? new Date(b.message.created_at).getTime() : b.entry.ts;
    if (ta !== tb) {
      return ta - tb;
    }
    if (a.type === b.type) {
      return 0;
    }
    return a.type === 'message' ? -1 : 1;
  });
  return rows;
});

async function scrollChatToBottom(): Promise<void> {
  await nextTick();
  const el = chatEl.value;
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
}

watch(
  () => timelineRows.value.length,
  () => {
    void scrollChatToBottom();
  },
);

watch(
  () => messages.value.length,
  () => {
    void scrollChatToBottom();
  },
);

function validConversationId(c: AiConversation): c is AiConversation & { id: number } {
  return typeof c.id === 'number' && Number.isFinite(c.id);
}

/** Active conversation id, or null if none / invalid (never undefined). */
function resolvedActiveId(): number | null {
  const v = activeId.value;
  if (v === null || v === undefined || !Number.isFinite(v)) {
    return null;
  }
  return v;
}

const hasActiveConversation = computed(() => resolvedActiveId() !== null);

async function loadConversations(): Promise<void> {
  const list = await DefaultService.listAiConversations();
  conversations.value = list.filter(validConversationId);
  if (resolvedActiveId() === null && conversations.value.length) {
    activeId.value = conversations.value[0].id;
  } else if (resolvedActiveId() !== null && !conversations.value.some((c) => c.id === activeId.value)) {
    activeId.value = conversations.value[0]?.id ?? null;
  }
}

async function loadMessages(): Promise<void> {
  const cid = activeId.value;
  if (cid === null || cid === undefined || !Number.isFinite(cid)) {
    messages.value = [];
    return;
  }
  messages.value = await DefaultService.listAiMessages({ conversationId: cid });
}

async function onNewConversation(): Promise<void> {
  try {
    const c = await DefaultService.createAiConversation({});
    if (!validConversationId(c)) {
      notify({ id: 'ai-new', type: 'error', title: 'AI', description: 'Invalid conversation from server.' });
      return;
    }
    await loadConversations();
    activeId.value = c.id;
    await loadMessages();
  } catch {
    notify({ id: 'ai-new', type: 'error', title: 'AI', description: 'Could not create conversation.' });
  }
}

async function onDeleteConversation(): Promise<void> {
  const id = resolvedActiveId();
  if (id === null) {
    return;
  }
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

function onComposerKeydown(e: KeyboardEvent): void {
  if (e.key !== 'Enter' || e.shiftKey) {
    return;
  }
  e.preventDefault();
  void onSend();
}

async function onSend(): Promise<void> {
  const text = draft.value.trim();
  if (!text || sending.value) {
    return;
  }
  sending.value = true;
  try {
    let cid = resolvedActiveId();
    if (cid === null) {
      const c = await DefaultService.createAiConversation({});
      if (!validConversationId(c)) {
        notify({ id: 'ai-new', type: 'error', title: 'AI', description: 'Invalid conversation from server.' });
        return;
      }
      await loadConversations();
      activeId.value = c.id;
      cid = c.id;
    }
    await DefaultService.createAiMessage({
      conversationId: cid,
      requestBody: { content: text },
    });
    agentProcessing.value = true;
    draft.value = '';
    await loadMessages();
    await scrollChatToBottom();
  } catch (e: unknown) {
    notifyApiError(e, {
      id: 'ai-send',
      title: 'AI',
      fallbackMessage: 'Could not send message.',
    });
  } finally {
    sending.value = false;
  }
}

async function onStop(): Promise<void> {
  const cid = resolvedActiveId();
  if (cid === null) {
    return;
  }
  try {
    await DefaultService.stopAiAgent({ conversationId: cid });
    pushActivity('Stop requested', 'stop');
  } catch {
    notify({ id: 'ai-stop', type: 'error', title: 'AI', description: 'Could not stop agent.' });
  }
}

async function onConfirm(approve: boolean): Promise<void> {
  const cid = resolvedActiveId();
  if (!pending.value || cid === null) {
    return;
  }
  try {
    await DefaultService.confirmAiTool({
      conversationId: cid,
      requestBody: { tool_call_id: pending.value.tool_call_id, approve },
    });
    pending.value = null;
    if (approve) {
      agentProcessing.value = true;
    }
    await loadMessages();
  } catch (e: unknown) {
    notifyApiError(e, {
      id: 'ai-confirm',
      title: 'AI',
      fallbackMessage: 'Could not confirm tool.',
    });
  }
}

watch(activeId, (newId, oldId) => {
  pending.value = null;
  activityLog.value = [];
  if (oldId != null && newId !== oldId) {
    agentProcessing.value = false;
  }
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
    const cur = resolvedActiveId();
    if (cur !== null && Number.isFinite(cid) && cid !== cur) {
      return;
    }
    const forActiveConv = cur !== null && Number.isFinite(cid) && cid === cur;
    const kind = String(last.kind ?? '');
    if (forActiveConv) {
      if (kind === 'needs_confirmation') {
        agentProcessing.value = false;
      } else if (kind === 'done' || kind === 'error') {
        agentProcessing.value = false;
      } else if (kind === 'llm_request' || kind === 'tool_attempt' || kind === 'tool_result') {
        agentProcessing.value = true;
      }
    }
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
    const logKind =
      kind === 'tool_attempt' || kind === 'tool_result' || kind === 'error' || kind === 'needs_confirmation';
    if (logKind) {
      pushActivity(describeAgentEvent(last as Record<string, unknown>), kind);
    }
  },
  { deep: true },
);

onMounted(async () => {
  try {
    await loadConversations();
    await loadMessages();
    await scrollChatToBottom();
  } catch {
    notify({ id: 'ai-load', type: 'error', title: 'AI', description: 'Failed to load (admin only?).' });
  }
});

function convLabel(c: AiConversation): string {
  if (c.title) {
    return c.title;
  }
  if (typeof c.id === 'number' && Number.isFinite(c.id)) {
    return `Conversation #${c.id}`;
  }
  return 'Conversation';
}

function formatMeta(m: Record<string, unknown>): string {
  if (!m || typeof m !== 'object') {
    return '';
  }
  if (m.error) {
    return JSON.stringify(m);
  }
  if (m.system) {
    return '[system]';
  }
  return '';
}

function roleLabel(role: AiMessage.role): string {
  if (role === AiMessage.role.USER) {
    return 'You';
  }
  if (role === AiMessage.role.ASSISTANT) {
    return 'Assistant';
  }
  return 'Tool';
}

function messageBody(m: AiMessage): string {
  const meta = formatMeta(m.metadata);
  if (!meta) {
    return m.content;
  }
  return `${m.content}\n${meta}`;
}

function activityLineClass(kind?: string): string {
  if (kind === 'error') {
    return 'ai-agent-line ai-agent-line--err';
  }
  if (kind === 'needs_confirmation') {
    return 'ai-agent-line ai-agent-line--confirm';
  }
  if (kind === 'tool_result') {
    return 'ai-agent-line ai-agent-line--ok';
  }
  return 'ai-agent-line';
}

function formatShortTime(ts: number): string {
  return new Date(ts).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}
</script>

<template>
  <div class="ai-wrap">
    <div class="ai-layout">
      <aside class="ai-sidebar">
        <div class="ai-sidebar-head">
          <h2 class="ai-sidebar-title">Conversations</h2>
          <button type="button" class="btn-ghost ai-sidebar-new" @click="onNewConversation">New</button>
        </div>
        <ul class="ai-conv-list">
          <li v-for="c in conversations" :key="String(c.id)">
            <button
              type="button"
              class="ai-conv-btn"
              :class="{ active: resolvedActiveId() === c.id }"
              @click="activeId = c.id"
            >
              <span class="ai-conv-label">{{ convLabel(c) }}</span>
              <span class="ai-conv-id">#{{ c.id }}</span>
            </button>
          </li>
        </ul>
        <p v-if="!conversations.length" class="ai-sidebar-empty muted">No chats yet — send a message or click New to start.</p>
        <button
          type="button"
          class="btn-danger-ghost ai-sidebar-del"
          :disabled="!hasActiveConversation"
          @click="onDeleteConversation"
        >
          Delete conversation
        </button>
      </aside>

      <section class="ai-chat">
        <div class="pane-head ai-chat-head">
          <div class="ai-chat-head-main">
            <span class="ai-chat-title">AI assistant</span>
            <span
              v-if="showAgentSpinner"
              class="ai-agent-spinner"
              aria-label="Agent working"
              role="status"
            />
            <span v-if="resolvedActiveId() != null" class="ai-chat-session">#{{ resolvedActiveId() }}</span>
            <span v-else class="muted">—</span>
          </div>
          <button type="button" class="btn-ghost ai-stop-btn" @click="onStop">Stop agent</button>
        </div>

        <ul ref="chatEl" class="lines">
          <li v-if="!timelineRows.length" class="ai-empty-line muted">
            No messages yet. Configure the model in Settings → AI, then send a prompt.
          </li>
          <template v-for="(row, idx) in timelineRows" v-else :key="row.type === 'message' ? `m-${row.message.id}` : `a-${row.entry.ts}-${idx}`">
            <ChatMessageLine
              v-if="row.type === 'message'"
              :user="roleLabel(row.message.role)"
              :message="messageBody(row.message)"
              :keyword="false"
              :from-sent="row.message.role === AiMessage.role.USER"
              :user-marked="row.message.role === AiMessage.role.TOOL"
              :user-is-sus="false"
              :suspicious-title="''"
              :first-message="false"
              :badge-tags="[]"
              :show-timestamp="false"
              :created-at="row.message.created_at"
              :preserve-line-breaks="true"
            />
            <li v-else :class="activityLineClass(row.entry.kind)">
              <span class="ts">{{ formatShortTime(row.entry.ts) }}</span>
              <span class="ai-nick">agent</span>
              <span class="txt">{{ row.entry.text }}</span>
            </li>
          </template>
        </ul>

        <div v-if="pending" class="ai-pending panel">
          <p class="ai-pending-title">
            <strong>{{ pending.tool_name }}</strong>
            <span class="muted"> needs confirmation</span>
          </p>
          <pre v-if="pending.arguments.trim()" class="ai-args">{{ pending.arguments }}</pre>
          <div class="row">
            <Button native-type="button" class="btn-primary" @click="onConfirm(true)">Approve</Button>
            <Button native-type="button" variant="ghost" class="btn-ghost" @click="onConfirm(false)">Reject</Button>
          </div>
        </div>

        <form class="composer" @submit.prevent="onSend">
          <label class="ai-compose-label">
            <span>Message</span>
            <textarea
              v-model="draft"
              class="composer-textarea"
              rows="3"
              name="ai_message"
              autocomplete="off"
              autocorrect="off"
              autocapitalize="off"
              spellcheck="false"
              placeholder="Message the agent… (Enter to send, Shift+Enter for newline)"
              @keydown="onComposerKeydown"
            />
          </label>
          <Button
            class="btn-send"
            native-type="submit"
            :loading="sending"
            :disabled="!draft.trim() || sending"
            full-width
          >
            {{ sending ? 'Sending…' : 'Send' }}
          </Button>
        </form>
      </section>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ai-wrap {
  padding: 0.75rem 1.25rem;
  max-width: min(1680px, 96vw);
  width: 100%;
  margin: 0 auto;
  box-sizing: border-box;
  min-height: calc(100vh - 4rem);
}

.ai-layout {
  display: grid;
  grid-template-columns: minmax(260px, 300px) 1fr;
  gap: 1rem;
  align-items: stretch;
  min-height: min(78vh, 720px);
}

.ai-sidebar {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  padding: 0.65rem 0.5rem;
  background: linear-gradient(165deg, rgba(145, 71, 255, 0.06) 0%, var(--bg-elevated) 42%);
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.04) inset;
}

.ai-sidebar-title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--text-muted);
  text-transform: uppercase;
}

.ai-sidebar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.15rem 0.25rem 0.35rem;
  border-bottom: 1px solid var(--border);
}

.ai-sidebar-new {
  font-size: 0.78rem;
  font-weight: 600;
}

.ai-sidebar-empty {
  margin: 0.25rem 0.35rem;
  font-size: 0.78rem;
  line-height: 1.35;
}

.ai-conv-list {
  list-style: none;
  margin: 0;
  padding: 0 0.15rem;
  flex: 1;
  overflow: auto;
  min-height: 0;
  max-height: min(52vh, 28rem);
}

.ai-conv-btn {
  width: 100%;
  text-align: left;
  padding: 0.45rem 0.5rem;
  border-radius: 0.3rem;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font: inherit;
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  transition:
    background 0.12s ease,
    border-color 0.12s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.04);
  }

  &.active {
    border-color: rgba(145, 71, 255, 0.45);
    background: rgba(145, 71, 255, 0.1);
    box-shadow: 0 0 0 1px rgba(145, 71, 255, 0.12);
  }
}

.ai-conv-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.82rem;
  font-weight: 600;
}

.ai-conv-id {
  flex-shrink: 0;
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

.btn-danger-ghost {
  border: 1px solid var(--danger, #c44);
  color: var(--danger, #c44);
  background: transparent;
  border-radius: 0.3rem;
  padding: 0.4rem 0.5rem;
  cursor: pointer;
  font: inherit;
  font-size: 0.78rem;
  margin-top: auto;

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
}

.ai-sidebar-del {
  flex-shrink: 0;
}

/* Match WatchChatSection: chat column */
.ai-chat {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  overflow: hidden;
}

.pane-head {
  flex-shrink: 0;
  min-height: 2.875rem;
  display: flex;
  align-items: center;
  box-sizing: border-box;
  padding: 0.35rem 0;
}

.ai-chat-head {
  width: 100%;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  border-bottom: 1px solid var(--border);
  margin: 0;
  padding-left: 0.65rem;
  padding-right: 0.65rem;
  font-size: 0.85rem;
  font-weight: 600;
}

.ai-chat-head-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem;
  min-width: 0;
}

.ai-agent-spinner {
  flex-shrink: 0;
  width: 0.85rem;
  height: 0.85rem;
  border: 2px solid rgba(255, 255, 255, 0.18);
  border-top-color: var(--accent-bright);
  border-radius: 50%;
  animation: ai-agent-spin 0.65s linear infinite;
}

@keyframes ai-agent-spin {
  to {
    transform: rotate(360deg);
  }
}

.ai-chat-title {
  flex-shrink: 0;
  line-height: 1.25;
  letter-spacing: 0.02em;
}

.ai-chat-session {
  font-weight: 700;
  color: var(--accent-bright);
  letter-spacing: 0.02em;
  font-variant-numeric: tabular-nums;
}

.ai-stop-btn {
  flex-shrink: 0;
  font-size: 0.78rem;
  font-weight: 600;
}

.lines {
  list-style: none;
  margin: 0;
  padding: 0.4rem;
  flex: 1 1 0;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
  font-size: 0.82rem;
  line-height: 1.35;
}

.ai-empty-line {
  list-style: none;
  padding: 1.25rem 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  text-align: center;
  font-size: 0.82rem;
  line-height: 1.45;
}

/* Agent activity rows: same rhythm as ChatMessageLine */
.ai-agent-line {
  list-style: none;
  padding: 0.2rem 0.15rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.15rem 0.35rem;
  font-size: 0.82rem;
  line-height: 1.35;

  .ts {
    display: inline-block;
    min-width: 6.75rem;
    margin-right: 0.15rem;
    font-size: 0.72rem;
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .ai-nick {
    font-weight: 700;
    color: #c4a8ff;
    margin-right: 0.35rem;
    flex-shrink: 0;
  }

  .txt {
    word-break: break-word;
    color: var(--text-muted);
    flex: 1 1 12rem;
    min-width: 0;
  }

  &--ok {
    background: rgba(0, 245, 147, 0.05);
  }

  &--confirm {
    background: rgba(255, 193, 7, 0.08);
    border-left: 2px solid rgba(255, 193, 7, 0.45);

    .txt {
      color: var(--text);
    }
  }

  &--err {
    background: rgba(244, 67, 54, 0.08);
    border-left: 2px solid rgba(244, 67, 54, 0.5);

    .txt {
      color: #f8a8a0;
    }
  }
}

.panel {
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  padding: 0.65rem 0.75rem;
  background: rgba(145, 71, 255, 0.06);
  margin: 0 0.5rem 0.35rem;
  flex-shrink: 0;
}

.ai-pending {
  border-color: rgba(145, 71, 255, 0.45);
  box-shadow: 0 0 0 1px rgba(145, 71, 255, 0.12);
}

.ai-pending-title {
  margin: 0 0 0.35rem;
  font-size: 0.85rem;
}

.ai-args {
  font-size: 0.76rem;
  overflow: auto;
  max-height: 8rem;
  margin: 0.35rem 0 0.5rem;
  padding: 0.4rem 0.5rem;
  border-radius: 0.25rem;
  background: var(--bg-base);
  border: 1px solid var(--border);
  white-space: pre-wrap;
  word-break: break-word;
}

.composer {
  border-top: 1px solid var(--border);
  padding: 0.5rem;
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.4rem;
  font-size: 0.78rem;
  flex-shrink: 0;

  .ai-compose-label {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    color: var(--text-muted);
    margin: 0;

    span {
      font-size: 0.72rem;
    }
  }

  select,
  input,
  textarea {
    padding: 0.35rem 0.4rem;
    border-radius: 0.2rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.85rem;
  }

  .composer-textarea {
    resize: vertical;
    min-height: 3.25rem;
    line-height: 1.35;
    font-family: inherit;
    width: 100%;
    box-sizing: border-box;
  }

  .btn-send {
    grid-column: 1 / -1;
    padding: 0.45rem;
    border: none;
    border-radius: 0.25rem;
    background: var(--accent);
    color: #fff;
    font-weight: 600;
    cursor: pointer;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }
}

.row {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.btn-primary {
  border: none;
  border-radius: 0.3rem;
  padding: 0.4rem 0.75rem;
  background: var(--accent);
  color: var(--bg-base, #0d0d0f);
  cursor: pointer;
  font: inherit;
  font-weight: 600;
}

@media (max-width: 800px) {
  .ai-layout {
    grid-template-columns: 1fr;
    min-height: unset;
  }

  .ai-conv-list {
    max-height: 14rem;
  }
}
</style>
