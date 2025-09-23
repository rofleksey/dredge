<template>
  <div class="send-message-container">
    <div class="send-message-input">
      <input
          type="text"
          :value="messageText"
          @input="$emit('update:messageText', ($event.target as HTMLInputElement).value)"
          placeholder="Type your message..."
          @keyup.enter="$emit('send-message')"
      >
      <button @click="$emit('send-message')">
        <i class="material-icons">send</i>
      </button>
    </div>
    <div class="channel-info">
      <span v-if="selectedUsername">Sending as: {{ selectedUsername }}</span>
      <span v-if="currentChannel">Channel: {{ currentChannel }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  selectedUsername: string | null
  currentChannel: string | null
  messageText: string
}>()

defineEmits<{
  'update:messageText': [value: string]
  'send-message': []
}>()
</script>

<style scoped>
.send-message-container {
  background-color: var(--surface-color);
  border-radius: 8px;
  padding: 1rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.send-message-input {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.send-message-input input {
  flex: 1;
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 0.75rem;
  color: var(--text-primary);
  font-size: 0.9rem;
}

.send-message-input input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.send-message-input button {
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  padding: 0.75rem 1rem;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.send-message-input button:hover {
  background-color: var(--primary-dark);
}

.channel-info {
  font-size: 0.8rem;
  color: var(--text-secondary);
  display: flex;
  justify-content: space-between;
}
</style>