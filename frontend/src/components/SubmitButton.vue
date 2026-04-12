<script setup lang="ts">
withDefaults(
  defineProps<{
    loading?: boolean;
    nativeType?: 'submit' | 'button';
    disabled?: boolean;
  }>(),
  {
    loading: false,
    nativeType: 'submit',
    disabled: false,
  },
);
</script>

<template>
  <button
    :type="nativeType"
    :disabled="loading || disabled"
    :aria-busy="loading"
    class="submit-btn"
  >
    <span v-if="loading" class="btn-submit-spinner" aria-hidden="true" />
    <span class="submit-btn-label"><slot /></span>
  </button>
</template>

<style scoped lang="scss">
.submit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
}
</style>
