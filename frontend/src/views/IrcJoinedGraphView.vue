<script setup lang="ts">
import * as Plot from '@observablehq/plot';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { ApiError, DefaultService } from '../api/generated';
import type { IrcJoinedSample } from '../api/generated';
import { notify } from '../lib/notify';

defineOptions({ name: 'IrcJoinedGraphView' });

const ircJoinLimit = 100;
const historyDays = 7;

const loading = ref(true);
const samples = ref<IrcJoinedSample[]>([]);
const chartEl = ref<HTMLDivElement | null>(null);

const points = computed(() => {
  const rows = samples.value
    .map((s) => ({
      t: new Date(s.captured_at),
      v: s.joined_count,
    }))
    .filter((p) => Number.isFinite(p.t.getTime()));

  rows.sort((a, b) => a.t.getTime() - b.t.getTime());

  return rows;
});

function segmentOver(y1: number, y2: number): boolean {
  return Math.max(y1, y2) > ircJoinLimit;
}

type JoinedEdge = { t1: Date; v1: number; t2: Date; v2: number; over: boolean };

function disposeChart(): void {
  chartEl.value?.replaceChildren();
}

function renderChart(): void {
  if (!chartEl.value) {
    return;
  }

  disposeChart();

  const pts = points.value;
  if (pts.length === 0) {
    return;
  }

  const w = chartEl.value.clientWidth;
  const h = chartEl.value.clientHeight;
  if (w < 48 || h < 48) {
    return;
  }

  const edges: JoinedEdge[] = [];
  for (let i = 0; i < pts.length - 1; i++) {
    const a = pts[i];
    const b = pts[i + 1];
    edges.push({
      t1: a.t,
      v1: a.v,
      t2: b.t,
      v2: b.v,
      over: segmentOver(a.v, b.v),
    });
  }

  const maxY = Math.max(ircJoinLimit, ...pts.map((p) => p.v), 1);
  const minY = 0;

  const accent = 'rgba(145, 71, 255, 0.9)';
  const overColor = '#ff6b6b';

  const figure = Plot.plot({
    width: w,
    height: h,
    marginLeft: 44,
    marginRight: 12,
    marginTop: 8,
    marginBottom: 36,
    style: {
      color: 'var(--text-muted)',
      fontSize: '11px',
      background: 'transparent',
    },
    x: { type: 'utc', label: 'Time', grid: true },
    y: {
      label: 'Joined',
      grid: true,
      domain: [minY, maxY + Math.max(2, Math.ceil(maxY * 0.06))],
    },
    marks: [
      Plot.ruleY([ircJoinLimit], {
        stroke: 'rgba(255, 255, 255, 0.28)',
        strokeWidth: 1,
        strokeDasharray: '5 4',
      }),
      Plot.link(edges, {
        x1: 't1',
        y1: 'v1',
        x2: 't2',
        y2: 'v2',
        stroke: (d: JoinedEdge) => (d.over ? overColor : accent),
        strokeWidth: 2,
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
      }),
      Plot.dot(pts, {
        x: 't',
        y: 'v',
        r: 2.5,
        fill: (d: (typeof pts)[number]) => (d.v > ircJoinLimit ? overColor : accent),
        stroke: 'rgba(0,0,0,0.35)',
        strokeWidth: 0.5,
        title: (d: (typeof pts)[number]) => `${d.v} joined · ${d.t.toLocaleString()}`,
      }),
      Plot.axisX({ fontSize: 10, tickFormat: '%b %d %H:%M' }),
      Plot.axisY({ fontSize: 10, ticks: 6 }),
      Plot.gridX({ stroke: 'rgba(255, 255, 255, 0.06)' }),
      Plot.gridY({ stroke: 'rgba(255, 255, 255, 0.06)' }),
    ],
  });

  chartEl.value.appendChild(figure);
}

async function load(): Promise<void> {
  loading.value = true;
  try {
    samples.value = await DefaultService.listIrcMonitorJoinedHistory({ days: historyDays });
  } catch (e: unknown) {
    samples.value = [];
    const msg =
      e instanceof ApiError && e.body && typeof e.body === 'object' && 'message' in e.body
        ? String((e.body as { message: string }).message)
        : 'Could not load IRC joined history.';
    notify({
      id: 'irc-joined-history',
      type: 'error',
      title: 'IRC joined',
      description: msg,
    });
  } finally {
    loading.value = false;
  }
}

function onResize(): void {
  if (!loading.value && points.value.length > 0) {
    renderChart();
  }
}

watch([loading, points], async () => {
  if (!loading.value) {
    await nextTick();
    renderChart();
  }
});

onMounted(async () => {
  window.addEventListener('resize', onResize);
  await load();
  await nextTick();
  renderChart();
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize);
  disposeChart();
});
</script>

<template>
  <div class="page">
    <header class="head">
      <h1>IRC joined (last {{ historyDays }} days)</h1>
      <p class="muted small">
        Sampled every 30 minutes. Line at {{ ircJoinLimit }}; segments above {{ ircJoinLimit }} are red.
      </p>
    </header>

    <p v-if="loading" class="muted">Loading…</p>
    <p v-else-if="!samples.length" class="muted">No samples yet — data appears after the server records snapshots.</p>
    <div v-show="!loading && samples.length > 0" ref="chartEl" class="chart-host" />
  </div>
</template>

<style scoped lang="scss">
.page {
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex: 1;
}

.head {
  margin-bottom: 0.65rem;
}

h1 {
  margin: 0 0 0.35rem;
  font-size: 1.15rem;
  color: var(--accent-bright);
}

.muted {
  color: var(--text-muted);
}

.small {
  font-size: 0.78rem;
  margin: 0;
}

.chart-host {
  width: 100%;
  min-height: 280px;
  height: min(42vh, 420px);
  max-width: 960px;
}
</style>
