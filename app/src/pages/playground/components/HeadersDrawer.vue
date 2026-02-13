<template>
  <v-navigation-drawer
    v-model="store.headersDrawerOpen"
    location="right"
    width="380"
    temporary
    color="background"
  >
    <div class="pa-4">
      <p class="text-h6 mb-4">Agent Headers</p>
      <p class="text-body-2 text-medium-emphasis mb-4">
        Headers sent with every request to the agent (e.g. Authorization, X-API-Key).
      </p>

      <!-- Preset shortcuts -->
      <div class="d-flex flex-wrap gap-2 mb-4">
        <v-chip
          v-for="key in headersStore.presetKeys"
          :key="key"
          variant="outlined"
          size="small"
          class="cursor-pointer"
          @click="headersStore.addPreset(key)"
        >
          {{ key }}
        </v-chip>
      </div>

      <!-- Custom headers list -->
      <div class="headers-list">
        <div
          v-for="(entry, index) in headersStore.headers"
          :key="index"
          class="d-flex gap-2 align-center mb-3"
        >
          <v-text-field
            :model-value="entry.key"
            label="Name"
            density="compact"
            hide-details
            variant="outlined"
            placeholder="Authorization"
            class="flex-grow-1"
            @update:model-value="(v) => headersStore.updateHeader(index, { key: v })"
          />
          <v-text-field
            :model-value="entry.value"
            label="Value"
            density="compact"
            hide-details
            variant="outlined"
            placeholder="Bearer token..."
            type="password"
            class="flex-grow-1"
            @update:model-value="(v) => headersStore.updateHeader(index, { value: v })"
          />
          <v-btn
            icon
            size="small"
            variant="text"
            color="error"
            @click="headersStore.removeHeader(index)"
          >
            <v-icon size="18">delete</v-icon>
          </v-btn>
        </div>
      </div>

      <v-btn
        block
        variant="tonal"
        prepend-icon="add"
        class="mt-2"
        @click="headersStore.addHeader()"
      >
        Add header
      </v-btn>
    </div>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
  import { useAgentPlaygroundStore } from '@/pages/playground/store/agentPlayground'
  import { useAgentHeadersStore } from '@/store/agentHeaders'

  const store = useAgentPlaygroundStore()
  const headersStore = useAgentHeadersStore()
</script>
