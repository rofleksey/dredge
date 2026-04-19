<script setup lang="ts">
import AppModal from '../../components/AppModal.vue';

defineProps<{
  open: boolean;
}>();

const manualChannel = defineModel<string>('manualChannel', { required: true });

const emit = defineEmits<{
  close: [];
  submit: [];
}>();
</script>

<template>
  <AppModal :open="open" title="Go to channel" @close="emit('close')">
    <form class="channel-entry-form" @submit.prevent="emit('submit')">
      <label class="channel-entry-label">
        <span>Twitch username</span>
        <input
          v-model="manualChannel"
          type="text"
          name="channel_login"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="off"
          spellcheck="false"
          placeholder="channel_name"
        />
      </label>
      <div class="channel-entry-actions">
        <button type="button" class="btn-cancel" @click="emit('close')">Cancel</button>
        <button type="submit" class="btn-confirm">Go</button>
      </div>
    </form>
  </AppModal>
</template>

<style scoped lang="scss">
.channel-entry-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.channel-entry-label {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  margin: 0;
  font-size: 0.82rem;
  color: var(--text-muted);

  input {
    padding: 0.45rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid var(--border);
    background: var(--bg-base);
    color: var(--text);
    font-size: 0.9rem;
  }
}

.channel-entry-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: flex-end;
}

.btn-cancel {
  padding: 0.4rem 0.75rem;
  border-radius: 0.25rem;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text);
  font-size: 0.85rem;
  cursor: pointer;

  &:hover {
    background: var(--bg-hover);
  }
}

.btn-confirm {
  padding: 0.4rem 0.85rem;
  border-radius: 0.25rem;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;

  &:hover {
    filter: brightness(1.08);
  }
}
</style>
