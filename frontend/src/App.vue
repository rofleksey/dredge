<template>
  <div id="app" class="app-container">
    <header>
      <div class="header-content">
        <div class="logo">
          <i class="material-icons">chat</i>
          <span>Dredge</span>
        </div>
        <div class="user-selector">
          <i class="material-icons">person</i>
          <select v-model="selectedUserId">
            <option v-for="user in availableUsers" :key="user.id" :value="user.id">{{ user.username }}</option>
          </select>
        </div>
      </div>
    </header>

    <main>
      <section class="messages-section">
        <div class="search-container">
          <div class="search-input">
            <input type="text" v-model="searchQuery"
                   placeholder="Enter query"
                   @keyup.enter="searchMessages">
            <button @click="searchMessages">
              <i class="material-icons">search</i>
              Search
            </button>
          </div>
          <div class="query-examples">
            Query: "channel:$channel username:$username date:$start~$end $text"
          </div>
        </div>

        <div class="messages-container">
          <div class="messages-header">
            <h3>Messages</h3>
            <div v-if="isRealtime" class="realtime-indicator">
              <div class="live-dot"></div>
              <span>Live Updates</span>
            </div>
          </div>

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

          <div class="pagination">
            <button @click="prevPage" :disabled="currentPage === 1">
              <i class="material-icons">chevron_left</i>
            </button>
            <span class="pagination-info">Page {{ currentPage }} of {{ totalPages }}</span>
            <button @click="nextPage" :disabled="currentPage === totalPages">
              <i class="material-icons">chevron_right</i>
            </button>
          </div>
        </div>

        <div class="send-message-container">
          <div class="send-message-input">
            <input type="text" v-model="messageText" placeholder="Type your message..." @keyup.enter="onSendMessage">
            <button @click="onSendMessage">
              <i class="material-icons">send</i>
            </button>
          </div>
          <div class="channel-info">
            <span v-if="selectedUsername">Sending as: {{ selectedUsername }}</span>
            <span v-if="currentChannel">Channel: {{ currentChannel }}</span>
          </div>
        </div>
      </section>

      <section class="stream-section">
        <div class="stream-container">
          <div class="stream-header">
            <h3>Twitch Stream</h3>
            <span v-if="currentChannel">#{{ currentChannel }}</span>
          </div>
          <div class="stream-iframe">
            <div v-if="!currentChannel" class="empty-state">
              <i class="material-icons">live_tv</i>
              <p>No channel selected</p>
            </div>
            <iframe
                v-else
                :src="`https://player.twitch.tv/?channel=${currentChannel}&parent=${FRAME_PARENT}&muted=true`"
                width="100%"
                height="100%"
                allowfullscreen>
            </iframe>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue'
import {useApiClient} from './lib/api/api'
import type {Message} from './lib/oapi'

const pageSize = 13

// @ts-ignore
const FRAME_PARENT = encodeURIComponent(process.env.NODE_ENV === 'development' ? "localhost" : window.location.hostname)

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

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

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
      currentChannel.value = messages.value[messages.value.length - 1]!.channel
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
      currentChannel.value = messages.value[messages.value.length - 1]!.channel
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

header {
  background-color: var(--surface-color);
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 700;
  font-size: 1.25rem;
}

.logo i {
  color: var(--primary-color);
}

.user-selector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.material-icons {
  font-size: 1.25rem;
}

select {
  background-color: var(--surface-color);
  color: var(--text-primary);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 0.5rem;
  font-size: 0.9rem;
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

.messages-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.stream-section {
  width: 100%;
  max-width: 400px;
}

@media (min-width: 768px) {
  .stream-section {
    min-width: 300px;
  }
}

.search-container {
  background-color: var(--surface-color);
  border-radius: 8px;
  padding: 1rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.search-input {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.search-input input {
  flex: 1;
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 0.75rem;
  color: var(--text-primary);
  font-size: 0.9rem;
}

.search-input input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.search-input button {
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  padding: 0.75rem 1rem;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

button {
  display: flex;
  align-items: center;
  justify-content: center;
}

.search-input button:hover {
  background-color: var(--primary-dark);
}

.query-examples {
  font-size: 0.8rem;
  color: var(--text-secondary);
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

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 0.5rem;
  gap: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.pagination button {
  background-color: var(--surface-color);
  color: var(--text-primary);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 0.5rem 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.pagination button:hover:not(:disabled) {
  background-color: rgba(255, 255, 255, 0.1);
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination-info {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

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

.stream-container {
  background-color: var(--surface-color);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  height: fit-content;
}

.stream-header {
  padding: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stream-iframe {
  width: 100%;
  aspect-ratio: 16/9;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text-secondary);
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