<script setup lang="ts">
import { useScroll } from '@vueuse/core';
import { nextTick, ref, watch } from 'vue';
import SubmitButton from '../../components/SubmitButton.vue';
import ChatMessageLine from '../../components/ChatMessageLine.vue';
import ChatSystemLine from '../../components/ChatSystemLine.vue';
import type { TwitchAccount } from '../../api/generated';
import type { ChatRow } from './types';

const props = defineProps<{
  displayRows: ChatRow[];
  selectedChannel: string;
  ircChatConnected: boolean;
  normCh: (c: string) => string;
  accounts: TwitchAccount[];
  sendingChat: boolean;
  chatScrollSig: string;
  composerKeydown: (e: KeyboardEvent) => void;
  sendChat: () => void | Promise<void>;
}>();

const sendAccountId = defineModel<number | null>('sendAccountId', { required: true });
const sendText = defineModel<string>('sendText', { required: true });

const chatEl = ref<HTMLElement | null>(null);
useScroll(chatEl);

watch(
  () => props.chatScrollSig,
  async () => {
    await nextTick();
    const el = chatEl.value;
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  },
);
</script>

<template>
  <section class="chat">
    <div class="pane-head chat-head">
      <div class="chat-head-main">
        <span class="chat-title-main">Chat</span>
        <span v-if="selectedChannel" class="chat-channel-tag">#{{ normCh(selectedChannel) }}</span>
        <span v-else class="chat-channel-empty muted">—</span>
        <span
          v-if="selectedChannel"
          class="irc-link-dot"
          role="img"
          :class="{ 'irc-link-dot--on': ircChatConnected, 'irc-link-dot--off': !ircChatConnected }"
          :title="ircChatConnected ? 'IRC monitor joined this channel' : 'IRC monitor not in this channel'"
          :aria-label="
            ircChatConnected ? 'IRC monitor joined this channel' : 'IRC monitor not in this channel'
          "
        />
      </div>
    </div>
    <ul ref="chatEl" class="lines">
      <template v-for="row in displayRows" :key="row.kind === 'gap' ? row.key : row.line.key">
        <ChatSystemLine
          v-if="row.kind === 'gap'"
          variant="gap"
          user=""
          :text="row.label"
        />
        <ChatMessageLine
          v-else
          :user="row.line.user"
          :message="row.line.message"
          :keyword="row.line.keyword"
          :user-marked="row.line.userMarked"
          :user-is-sus="row.line.userIsSus"
          :suspicious-title="row.line.susTitle"
          :from-sent="row.line.fromSent"
          :first-message="row.line.firstMessage"
          :badge-tags="row.line.badgeTags"
          :show-timestamp="false"
          :created-at="row.line.createdAtIso"
          :chatter-user-id="row.line.chatterUserId ?? undefined"
          :highlight-channel="normCh(selectedChannel)"
        />
      </template>
    </ul>

    <div class="composer">
      <label>
        <span>Send as</span>
        <select v-model.number="sendAccountId">
          <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.username }}</option>
        </select>
      </label>
      <label>
        <span>Message</span>
        <textarea
          v-model="sendText"
          class="composer-textarea"
          maxlength="500"
          rows="3"
          name="chat_message"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="off"
          spellcheck="false"
          placeholder="Say something… (Enter to send, Shift+Enter for newline)"
          @keydown="composerKeydown"
        />
      </label>
      <SubmitButton
        native-type="button"
        class="btn-send"
        :loading="sendingChat"
        :disabled="!accounts.length"
        @click="sendChat"
      >
        {{ sendingChat ? 'Sending…' : 'Chat' }}
      </SubmitButton>
    </div>
  </section>
</template>

<style scoped lang="scss">
.pane-head {
  flex-shrink: 0;
  min-height: 2.875rem;
  display: flex;
  align-items: center;
  box-sizing: border-box;
  padding: 0.35rem 0;
}

.chat {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.35rem;
  overflow: hidden;

  @media (min-width: 900px) {
    flex: 0 0 min(400px, 38vw);
    align-self: stretch;
    max-width: 400px;
    min-height: 0;
    max-height: 100%;
  }
}

.chat-head {
  width: 100%;
  justify-content: flex-start;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--border);
  margin: 0;
  padding-left: 0.5rem;
  padding-right: 0.5rem;
  font-size: 0.85rem;
  font-weight: 600;
  box-sizing: border-box;
}

.chat-head-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  flex: 1;
}

.irc-link-dot {
  flex-shrink: 0;
  width: 0.55rem;
  height: 0.55rem;
  margin-left: auto;
  border-radius: 50%;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.25);

  &--on {
    background: #2ecc71;
  }

  &--off {
    background: #c0392b;
  }
}

.chat-title-main {
  flex-shrink: 0;
  line-height: 1.25;
}

.chat-channel-tag {
  font-weight: 700;
  color: var(--accent-bright);
  letter-spacing: 0.02em;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-channel-empty {
  font-size: 0.85rem;
}

.muted {
  color: var(--text-muted);
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

.composer {
  border-top: 1px solid var(--border);
  padding: 0.5rem;
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.4rem;
  font-size: 0.78rem;
  flex-shrink: 0;

  label {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    color: var(--text-muted);

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
</style>
