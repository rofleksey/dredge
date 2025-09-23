<template>
  <div class="messages-list">
    <div v-if="loading" class="loading">
      <i class="material-icons">refresh</i>
      <span>Loading messages...</span>
    </div>

    <div v-else-if="messages.length === 0" class="empty-state">
      <i class="material-icons">forum</i>
      <p>No messages found. Try a different search query.</p>
    </div>

    <div v-else>
      <div v-for="message in messages" :key="message.id" class="message">
        <div class="message-header">
          <div>
            <span class="message-username">{{ message.username }}</span>
            &nbsp;
            <span class="message-channel">#{{ message.channel }}</span>
          </div>

          <div class="message-timestamp">{{ formatDate(message.created) }}</div>
        </div>

        <div class="message-text">{{ message.text }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Message } from '../lib/oapi'

defineProps<{
  messages: Message[]
  loading: boolean
}>()

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}
</script>

<style scoped>
.messages-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.message {
  padding: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.message:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.message-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.message-username {
  font-weight: 500;
  color: var(--primary-color);
}

.message-channel {
  color: var(--text-secondary);
}

.message-text {
  font-size: 0.9rem;
  line-height: 1.4;
  text-align: left;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text-secondary);
  text-align: center;
}

.empty-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}
</style>