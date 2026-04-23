<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { RouterLink } from 'vue-router';
import { DefaultService } from '../api/generated';
import type { TwitchUser } from '../api/generated';
import { LoadMoreRow, PageHeader, TextInput } from '../components/core';
import { notifyApiError } from '../lib/notifyApiError';

defineOptions({ name: 'UsersView' });

const q = ref('');
const users = ref<TwitchUser[]>([]);
const loading = ref(false);
const loadingMore = ref(false);
const totalCount = ref<number | null>(null);
let debounceTimer: ReturnType<typeof setTimeout> | null = null;

async function load(append = false): Promise<void> {
  if (append) {
    loadingMore.value = true;
  } else {
    loading.value = true;
  }
  try {
    const last = append && users.value.length ? users.value[users.value.length - 1] : undefined;
    const [list, cnt] = await Promise.all([
      DefaultService.listTwitchDirectoryUsers({
        username: q.value.trim() || undefined,
        limit: 100,
        cursorId: append && last ? last.id : undefined,
      }),
      DefaultService.countTwitchDirectoryUsers({
        username: q.value.trim() || undefined,
      }),
    ]);
    if (append) {
      const seen = new Set(users.value.map((u) => u.id));
      for (const u of list) {
        if (!seen.has(u.id)) {
          users.value.push(u);
          seen.add(u.id);
        }
      }
    } else {
      users.value = list;
    }
    totalCount.value = cnt.total;
  } catch (e) {
    if (!append) {
      users.value = [];
    }
    totalCount.value = null;
    notifyApiError(e, {
      id: 'users-load',
      title: 'Users',
      fallbackMessage: 'Could not load users.',
    });
  } finally {
    loading.value = false;
    loadingMore.value = false;
  }
}

async function loadMore(): Promise<void> {
  if (!users.value.length || loadingMore.value || loading.value) {
    return;
  }
  await load(true);
}

onMounted(() => {
  void load();
});

watch(q, () => {
  if (debounceTimer) {
    clearTimeout(debounceTimer);
  }
  debounceTimer = setTimeout(() => {
    debounceTimer = null;
    void load(false);
  }, 300);
});
</script>

<template>
  <div class="page users-page">
    <PageHeader title="Users" :total-count="totalCount" layout="inline" />

    <div class="users-search">
      <TextInput
        v-model="q"
        label="Search by username"
        type="search"
        autocomplete="off"
        placeholder="Substring match…"
        density="compact"
        surface="base"
      />
    </div>

    <p v-if="loading" class="muted muted--body">Loading…</p>
    <ul v-else class="user-list">
      <li v-for="u in users" :key="u.id" :class="{ suspicious: u.is_sus }">
        <RouterLink
          class="user-link"
          :class="{ marked: u.marked, suspicious: u.is_sus }"
          :to="{ name: 'user', params: { id: String(u.id) } }"
        >
          {{ u.username }}
        </RouterLink>
        <span v-if="u.is_sus" class="tag tag-sus">suspicious</span>
        <span v-if="u.marked" class="tag tag-marked">marked</span>
        <span v-if="u.monitored" class="tag">monitored</span>
      </li>
    </ul>
    <LoadMoreRow v-if="!loading && users.length" :loading="loadingMore" @click="loadMore" />
    <p v-if="!loading && !users.length" class="muted muted--body">No users match.</p>
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.users-search {
  margin-bottom: 0.75rem;
  max-width: 24rem;
}

.user-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;

  li.suspicious {
    padding: 0.2rem 0.35rem;
    margin: 0 -0.35rem;
    border-radius: 0.25rem;
    background: rgba(220, 53, 69, 0.12);
    border: 1px solid rgba(220, 53, 69, 0.35);
  }
}

.user-link {
  color: var(--accent-bright);
  font-weight: 600;
  text-decoration: none;

  &:hover {
    text-decoration: underline;
  }

  &.marked {
    color: #ffc107;
  }

  &.suspicious {
    color: #ff6b7a;
  }
}

.tag {
  margin-left: 0.5rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.tag-marked {
  color: #ffc107;
}

.tag-sus {
  color: #ff6b7a;
}

</style>
