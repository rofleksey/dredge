<template>
  <div id="app" class="app-container">
    <AppHeader
        :available-users="availableUsers"
        v-model:selected-user-id="selectedUserId"
    />

    <main>
      <MessagesSection
          :messages="messages"
          :loading="loading"
          :is-realtime="isRealtime"
          :current-page="currentPage"
          :total-pages="totalPages"
          :selected-username="selectedUsername"
          :current-channel="currentChannel"
          @search="searchMessages"
          @prev-page="prevPage"
          @next-page="nextPage"
          @send-message="onSendMessage"
          v-model:search-query="searchQuery"
          v-model:message-text="messageText"
      />

      <StreamSection :current-channel="currentChannel" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useApiClient } from './lib/api/api'
import type { Message } from './lib/oapi'
import AppHeader from './components/AppHeader.vue'
import MessagesSection from './components/MessagesSection.vue'
import StreamSection from './components/StreamSection.vue'

const pageSize = 13

const apiClient = useApiClient()

const searchQuery = ref('')
const messages = ref<Message[]>([])
const currentPage = ref(1)
const totalPages = ref(1)
const loading = ref(false)
const isRealtime = ref(false)
const selectedUserId = ref(1)
const messageText = ref('')
const currentChannel = ref<string | null>(null)
const availableUsers = ref<Array<{ id: number, username: string }>>([])
const totalCount = ref(0)

const selectedUsername = computed(() => {
  const user = availableUsers.value.find(u => u.id === selectedUserId.value)
  return user?.username ?? null
})

const searchMessages = async () => {
  loading.value = true
  currentPage.value = 1

  try {
    const response = await apiClient.api.searchMessages({
      searchRequest: {
        query: searchQuery.value,
        offset: (currentPage.value - 1) * pageSize,
        limit: pageSize
      }
    })
    messages.value = response.data.messages
    totalCount.value = response.data.totalCount
    totalPages.value = Math.ceil(totalCount.value / pageSize)

    const queryParts = searchQuery.value.split(' ')
    const hasChannelFilter = queryParts.some(part => part.startsWith('channel:'))
    const hasOtherFilters = queryParts.some(part =>
        part.startsWith('username:') || part.startsWith('date:') ||
        (part && !part.startsWith('channel:'))
    )

    isRealtime.value = hasChannelFilter && !hasOtherFilters

    if (messages.value.length > 0) {
      currentChannel.value = messages.value[0]!.channel
    }
  } catch (error) {
    console.error('Error searching messages:', error)
    messages.value = []
  } finally {
    loading.value = false
  }
}

const loadMessages = async () => {
  loading.value = true
  try {
    const response = await apiClient.api.searchMessages({
      searchRequest: {
        query: searchQuery.value,
        offset: (currentPage.value - 1) * pageSize,
        limit: pageSize
      }
    })
    messages.value = response.data.messages
    totalCount.value = response.data.totalCount
    totalPages.value = Math.ceil(totalCount.value / pageSize)

    if (messages.value.length > 0) {
      currentChannel.value = messages.value[0]!.channel
    }
  } catch (error) {
    console.error('Error loading messages:', error)
  } finally {
    loading.value = false
  }
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loadMessages()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadMessages()
  }
}

function onSendMessage() {
  const channel = currentChannel.value
  const username = selectedUsername.value
  const text = messageText.value.trim()

  if (!channel || !username || !text) {
    return
  }

  sendMessage(channel, username, text).catch((e) => {
    console.error('Error sending message:', e)
  })
}

async function sendMessage(channel: string, username: string, text: string) {
  await apiClient.api.sendMessage({
    sendMessageRequest: {
      username,
      channel,
      text
    }
  })

  if (isRealtime.value) {
    await loadMessages()
  }
}

const loadUsers = async () => {
  try {
    const response = await apiClient.api.getUsers()
    availableUsers.value = response.data.usernames.map((username, index) => ({
      id: index + 1,
      username
    }))
  } catch (error) {
    console.error('Error loading users:', error)
  }
}

onMounted(async () => {
  await loadUsers()
  await loadMessages()
})
</script>

<style>
:root {
  --primary-color: #9147ff;
  --primary-dark: #772ce8;
  --surface-color: #1f1f23;
  --background-color: #0e0e10;
  --text-primary: #efeff1;
  --text-secondary: #adadb8;
  --error-color: #eb0400;
  --success-color: #00d95f;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Roboto', sans-serif;
  background-color: var(--background-color);
  color: var(--text-primary);
  line-height: 1.5;
}

.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  max-width: 1200px;
  margin: 0 auto;
}

main {
  display: flex;
  flex: 1;
  flex-direction: column;
  padding: 1rem;
  gap: 1rem;
}

@media (min-width: 768px) {
  main {
    flex-direction: row;
  }
}
</style>