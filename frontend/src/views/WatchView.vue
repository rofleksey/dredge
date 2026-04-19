<script setup lang="ts">
import TwitchPlayer from '../components/TwitchPlayer.vue';
import { useWatchView } from '../composables/useWatchView';
import WatchChannelEntryModal from './watch/WatchChannelEntryModal.vue';
import WatchChatSection from './watch/WatchChatSection.vue';
import WatchMonitoredSidebar from './watch/WatchMonitoredSidebar.vue';
import WatchStreamStrip from './watch/WatchStreamStrip.vue';
import WatchViewerModal from './watch/WatchViewerModal.vue';

defineOptions({ name: 'WatchView' });

const {
  twitchStore,
  selectedChannel,
  sendAccountId,
  sendText,
  channelLive,
  loadingChannelMeta,
  loadingMonitoredSidebar,
  viewerModalOpen,
  viewerChatters,
  loadingViewerChatters,
  viewerFilterQuery,
  viewerSort,
  channelEntryModalOpen,
  manualChannelInput,
  sendingChat,
  onlineMonitored,
  offlineMonitored,
  monitoredSidebar,
  ircChatConnected,
  displayedViewerChatters,
  displayRows,
  chatScrollSig,
  formatSessionClock,
  formatViewerDisplay,
  formatPresentElapsed,
  formatAccountDate,
  normCh,
  selectChannel,
  applyManualChannel,
  onComposerKeydown,
  sendChat,
} = useWatchView();
</script>

<template>
  <div class="watch">
    <div class="watch-layout">
      <WatchMonitoredSidebar
        :loading="loadingMonitoredSidebar"
        :online-monitored="onlineMonitored"
        :offline-monitored="offlineMonitored"
        :monitored-sidebar="monitoredSidebar"
        :selected-channel="selectedChannel"
        :norm-ch="normCh"
        @select="selectChannel"
        @open-channel="channelEntryModalOpen = true"
      />

      <div class="watch-main">
        <div class="grid">
          <section class="video">
            <TwitchPlayer :channel="selectedChannel" />

            <WatchStreamStrip
              :channel-live="channelLive"
              :loading-channel-meta="loadingChannelMeta"
              :selected-channel="selectedChannel"
              :format-session-clock="formatSessionClock"
              :format-viewer-display="formatViewerDisplay"
              @open-viewers="viewerModalOpen = true"
            />
          </section>

          <WatchChatSection
            v-model:send-account-id="sendAccountId"
            v-model:send-text="sendText"
            :display-rows="displayRows"
            :selected-channel="selectedChannel"
            :irc-chat-connected="ircChatConnected"
            :norm-ch="normCh"
            :accounts="twitchStore.accounts"
            :sending-chat="sendingChat"
            :chat-scroll-sig="chatScrollSig"
            :composer-keydown="onComposerKeydown"
            :send-chat="sendChat"
          />
        </div>
      </div>
    </div>

    <WatchChannelEntryModal
      v-model:manual-channel="manualChannelInput"
      :open="channelEntryModalOpen"
      @close="channelEntryModalOpen = false"
      @submit="applyManualChannel"
    />

    <WatchViewerModal
      v-model:viewer-filter-query="viewerFilterQuery"
      v-model:viewer-sort="viewerSort"
      :open="viewerModalOpen"
      :channel-live="channelLive"
      :format-session-clock="formatSessionClock"
      :format-viewer-display="formatViewerDisplay"
      :format-present-elapsed="formatPresentElapsed"
      :format-account-date="formatAccountDate"
      :viewer-chatters="viewerChatters"
      :loading-viewer-chatters="loadingViewerChatters"
      :displayed-viewer-chatters="displayedViewerChatters"
      :selected-channel="selectedChannel"
      :norm-ch="normCh"
      @close="viewerModalOpen = false"
    />
  </div>
</template>

<style scoped lang="scss">
.watch {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.75rem;
  flex: 1;
  min-height: 0;
}

.watch-layout {
  display: flex;
  flex: 1;
  min-height: 0;
  gap: 0.65rem;
  align-items: stretch;
}

.watch-main {
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  min-height: 0;
}

.grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  flex: 1;
  min-height: 0;
  overflow: hidden;

  @media (min-width: 900px) {
    flex-direction: row;
    align-items: stretch;
  }
}

.video {
  flex: 0 0 auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  min-height: 0;

  @media (min-width: 900px) {
    flex: 1 1 auto;
    min-width: 0;
  }
}

@media (max-width: 639px) {
  .watch-layout {
    flex-direction: column;
  }
}
</style>
