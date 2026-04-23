<script setup lang="ts">
import { computed, useId } from 'vue';

type InputDensity = 'comfortable' | 'compact';
type InputSurface = 'base' | 'elevated';

/**
 * Shared text input wrapper with optional label and error state.
 * Uses `v-model` via `modelValue`/`update:modelValue`.
 */
const props = withDefaults(
  defineProps<{
    modelValue: string | number;
    id?: string;
    label?: string;
    type?: string;
    autocomplete?: string;
    placeholder?: string;
    name?: string;
    required?: boolean;
    disabled?: boolean;
    min?: string | number;
    max?: string | number;
    step?: string | number;
    density?: InputDensity;
    surface?: InputSurface;
    error?: string;
  }>(),
  {
    id: '',
    label: '',
    type: 'text',
    autocomplete: '',
    placeholder: '',
    name: '',
    required: false,
    disabled: false,
    min: undefined,
    max: undefined,
    step: undefined,
    density: 'comfortable',
    surface: 'base',
    error: '',
  },
);

const emit = defineEmits<{ 'update:modelValue': [value: string] }>();

const generatedId = useId();
const inputId = computed(() => props.id || `core-input-${generatedId}`);
const describedBy = computed(() => (props.error ? `${inputId.value}-error` : undefined));

const controlClass = computed(() => {
  const list = [
    'text-input__control',
    props.density === 'comfortable' ? 'text-input__control--comfortable' : 'text-input__control--compact',
    props.surface === 'base' ? 'text-input__control--surface-base' : 'text-input__control--surface-elevated',
  ];
  if (props.error) {
    list.push('text-input__control--invalid');
  }
  return list;
});

function onInput(event: Event): void {
  const target = event.target as HTMLInputElement;
  emit('update:modelValue', target.value);
}
</script>

<template>
  <div class="text-input-root">
    <label v-if="label" :for="inputId" class="text-input-label">
      <span>{{ label }}</span>
      <input
        :id="inputId"
        :name="name || undefined"
        :type="type"
        :value="modelValue"
        :autocomplete="autocomplete || undefined"
        :placeholder="placeholder || undefined"
        :required="required"
        :disabled="disabled"
        :min="min"
        :max="max"
        :step="step"
        :aria-invalid="error ? 'true' : undefined"
        :aria-describedby="describedBy"
        :class="controlClass"
        @input="onInput"
        @change="onInput"
      />
    </label>
    <input
      v-else
      :id="inputId"
      :name="name || undefined"
      :type="type"
      :value="modelValue"
      :autocomplete="autocomplete || undefined"
      :placeholder="placeholder || undefined"
      :required="required"
      :disabled="disabled"
      :min="min"
      :max="max"
      :step="step"
      :aria-invalid="error ? 'true' : undefined"
      :aria-describedby="describedBy"
      :class="controlClass"
      @input="onInput"
      @change="onInput"
    />
    <p v-if="error" :id="`${inputId}-error`" role="alert" class="text-input-error">{{ error }}</p>
  </div>
</template>

<style scoped lang="scss">
.text-input-root {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.text-input-label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.text-input__control {
  border: 1px solid var(--border);
  border-radius: 0.25rem;
  color: var(--text);
  font-family: inherit;
}

.text-input__control--comfortable {
  padding: 0.55rem 0.65rem;
  font-size: 1rem;
}

.text-input__control--compact {
  padding: 0.4rem 0.55rem;
  font-size: 0.88rem;
}

.text-input__control--surface-base {
  background: var(--bg-base);
}

.text-input__control--surface-elevated {
  background: var(--bg-elevated);
}

.text-input__control--invalid {
  border-color: #ff6b6b;
}

.text-input-error {
  margin: 0;
  font-size: 0.78rem;
  color: #ff6b6b;
}
</style>
