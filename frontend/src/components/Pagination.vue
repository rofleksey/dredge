<template>
  <div class="pagination">
    <button @click="$emit('prev-page')" :disabled="currentPage === 1">
      <i class="material-icons">chevron_left</i>
    </button>
    <span class="pagination-info">Page {{ currentPage }} of {{ totalPages }}</span>
    <button @click="$emit('next-page')" :disabled="currentPage === totalPages">
      <i class="material-icons">chevron_right</i>
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  currentPage: number
  totalPages: number
}>()

defineEmits<{
  'prev-page': []
  'next-page': []
}>()
</script>

<style scoped>
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
</style>