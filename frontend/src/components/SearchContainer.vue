<template>
  <div class="search-container">
    <div class="search-input">
      <input
          type="text"
          :value="searchQuery"
          @input="$emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          placeholder="Enter query"
          @keyup.enter="$emit('search')"
      >
      <button @click="$emit('search')">
        <i class="material-icons">search</i>
        Search
      </button>
    </div>
    <div class="query-examples">
      Query: "channel:$channel username:$username date:$start~$end $text"
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  searchQuery: string
}>()

defineEmits<{
  'update:searchQuery': [value: string]
  search: []
}>()
</script>

<style scoped>
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
</style>