<template>
  <header>
    <div class="header-content">
      <div class="logo">
        <i class="material-icons">chat</i>
        <span>Dredge</span>
      </div>
      <div class="header-controls">
        <div class="user-selector">
          <i class="material-icons">person</i>
          <select :value="selectedUserId" @change="handleUserChange($event)">
            <option v-for="user in availableUsers" :key="user.id" :value="user.id">{{ user.username }}</option>
          </select>
        </div>
        <button class="logout-btn" @click="$emit('logout')">
          <i class="material-icons">logout</i>
          Logout
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  availableUsers: Array<{ id: number, username: string }>
  selectedUserId: number
}>()

const emit = defineEmits<{
  'update:selectedUserId': [value: number]
  logout: []
}>()

const handleUserChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  emit('update:selectedUserId', parseInt(target.value))
}
</script>

<style scoped>
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

.header-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
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

.logout-btn {
  background-color: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-secondary);
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s;
}

.logout-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}
</style>