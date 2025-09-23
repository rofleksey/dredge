<template>
  <div class="login-container">
    <div class="login-card">
      <div class="logo">
        <i class="material-icons">chat</i>
        <span>Dredge</span>
      </div>
      <form @submit.prevent="handleLogin" class="login-form">
        <div class="input-group">
          <label for="username">Username</label>
          <input
              id="username"
              type="text"
              v-model="username"
              required
              autocomplete="username"
          />
        </div>
        <div class="input-group">
          <label for="password">Password</label>
          <input
              id="password"
              type="password"
              v-model="password"
              required
              autocomplete="current-password"
          />
        </div>
        <button type="submit" :disabled="loading">
          <span v-if="loading">Logging in...</span>
          <span v-else>Login</span>
        </button>
        <div v-if="error" class="error-message">{{ error }}</div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useApiClient } from '../lib/api/api'
import { useAuthStore } from '../stores/auth-store'

const router = useRouter()
const authStore = useAuthStore()
const apiClient = useApiClient()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await apiClient.api.login({
      loginRequest: {
        username: username.value,
        password: password.value
      }
    })

    authStore.login(response.data.token)
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.msg || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: var(--background-color);
}

.login-card {
  background-color: var(--surface-color);
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  width: 100%;
  max-width: 400px;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 2rem;
}

.logo i {
  color: var(--primary-color);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.input-group label {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.input-group input {
  background-color: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 0.75rem;
  color: var(--text-primary);
  font-size: 0.9rem;
}

.input-group input:focus {
  outline: none;
  border-color: var(--primary-color);
}

button {
  background-color: var(--primary-color);
  color: white;
  border: none;
  border-radius: 4px;
  padding: 0.75rem;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

button:hover:not(:disabled) {
  background-color: var(--primary-dark);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  color: var(--error-color);
  font-size: 0.9rem;
  text-align: center;
}
</style>