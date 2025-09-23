import {defineStore} from 'pinia'
import {ref} from 'vue'

const localStorageKey = 'dredge-token'

export const useAuthStore = defineStore('auth', () => {
  const localToken = localStorage.getItem(localStorageKey)
  const token = ref<string | null>(localToken ? localToken : null)

  function login(newToken: string) {
    token.value = newToken
    localStorage.setItem(localStorageKey, newToken)
  }

  function logout() {
    token.value = null
    localStorage.removeItem(localStorageKey)
  }

  return {
    token,
    login,
    logout
  }
})
