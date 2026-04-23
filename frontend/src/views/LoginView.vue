<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Button, TextInput } from '../components/core';
import { notify } from '../lib/notify';
import { useAuthStore } from '../stores/auth';

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const email = ref('');
const password = ref('');
const loading = ref(false);

async function submit(): Promise<void> {
  loading.value = true;
  try {
    await auth.login(email.value.trim(), password.value);
    const redir = typeof route.query.redirect === 'string' ? route.query.redirect : '/';
    await router.push(redir);
  } catch {
    notify({
      id: 'login',
      type: 'error',
      title: 'Sign in failed',
      description: 'Invalid email or password.',
    });
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="page">
    <div class="card">
      <h1>Sign in</h1>
      <p class="hint">Use your Dredge account (admin for full access).</p>
      <form @submit.prevent="submit">
        <TextInput
          v-model="email"
          label="Email"
          type="email"
          required
          autocomplete="username"
          surface="base"
          density="comfortable"
        />
        <TextInput
          v-model="password"
          label="Password"
          type="password"
          required
          autocomplete="current-password"
          surface="base"
          density="comfortable"
        />
        <Button class="signin-btn" native-type="submit" :loading="loading" full-width>
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </Button>
      </form>
    </div>
  </div>
</template>

<style scoped lang="scss">
.page {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--bg-base);
}

.card {
  width: 100%;
  max-width: 380px;
  padding: 1.5rem;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 0.5rem;
}

h1 {
  margin: 0 0 0.35rem;
  font-size: 1.35rem;
}

.hint {
  margin: 0 0 1rem;
  font-size: 0.85rem;
  color: var(--text-muted);
}

form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.signin-btn {
  margin-top: 0.25rem;
  padding: 0.6rem;
}
</style>
