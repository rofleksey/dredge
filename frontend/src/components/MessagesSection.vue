<template>
  <section class="messages-section">
    <SearchContainer
        :search-query="searchQuery"
        @update:search-query="$emit('update:searchQuery', $event)"
        @search="$emit('search')"
    />

    <div class="messages-container">
      <div class="messages-header">
        <h3>Messages</h3>
        <div v-if="isRealtime" class="realtime-indicator">
          <div class="live-dot"></div>
          <span>Live Updates</span>
        </div>
      </div>

      <MessagesList
          :messages="messages"
          :loading="loading"
          @username-click="$emit('username-click', $event)"
          @channel-click="$emit('channel-click', $event)"
      />

      <Pagination
          :current-page="currentPage"
          :total-pages="totalPages"
          @prev-page="$emit('prev-page')"
          @next-page="$emit('next-page')"
      />
    </div>

    <SendMessageContainer
        :selected-username="selectedUsername"
        :current-channel="currentChannel"
        :message-text="messageText"
        @update:message-text="$emit('update:messageText', $event)"
        @send-message="$emit('send-message')"
    />
  </section>
</template>

<script setup lang="ts">
import SearchContainer from './SearchContainer.vue'
import MessagesList from './MessagesList.vue'
import Pagination from './Pagination.vue'
import SendMessageContainer from './SendMessageContainer.vue'
import type { Message } from '../lib/oapi'

defineProps<{
  messages: Message[]
  loading: boolean
  isRealtime: boolean
  currentPage: number
  totalPages: number
  selectedUsername: string | null
  currentChannel: string | null
  searchQuery: string
  messageText: string
}>()

defineEmits<{
  (e: 'update:searchQuery', value: string): void
  (e: 'update:messageText', value: string): void
  (e: 'search'): void
  (e: 'prev-page'): void
  (e: 'next-page'): void
  (e: 'send-message'): void
  (e: 'username-click', username: string): void
  (e: 'channel-click', channel: string): void
}>()
</script>

<style scoped>
.messages-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.messages-container {
  background-color: var(--surface-color);
  border-radius: 8px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.messages-header {
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.realtime-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  color: var(--success-color);
}

.live-dot {
  width: 8px;
  height: 8px;
  background-color: var(--success-color);
  border-radius: 50%;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
  100% {
    opacity: 1;
  }
}
</style>