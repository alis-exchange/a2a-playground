<template>
  <v-navigation-drawer
    :model-value="true"
    location="right"
    :width="drawerWidth"
    permanent
    class="playground-drawer"
  >
    <div class="pa-4">
      <div class="d-flex align-center justify-space-between mb-2">
        <span class="text-h6">Playground</span>
        <v-btn
          icon
          variant="text"
          size="small"
          @click="onClose"
        >
          <v-icon>close</v-icon>
        </v-btn>
      </div>
      <p class="text-body-2 text-medium-emphasis mb-4">Test and debug your agent connections</p>

      <v-divider class="mb-4" />

      <!-- Connection Settings -->
      <div class="mb-6">
        <div class="text-subtitle-2 mb-3">Connection Settings</div>

        <v-text-field
          :model-value="connectionStore.agentUrl"
          label="Agent URL"
          density="compact"
          variant="outlined"
          type="url"
          autocomplete="url"
          hide-details
          class="mb-3"
          @update:model-value="(v) => (connectionStore.agentUrl = v ?? '')"
        />

        <v-select
          :model-value="connectionStore.protocol"
          :items="transportItems"
          label="Transport"
          density="compact"
          variant="outlined"
          hide-details
          @update:model-value="(v) => (connectionStore.protocol = (v as 'grpc' | 'jsonrpc') ?? 'grpc')"
        />
      </div>

      <v-divider class="mb-4" />

      <!-- Authentication -->
      <div class="mb-6">
        <div class="text-subtitle-2 mb-2">Authentication</div>
        <p class="text-body-2 text-medium-emphasis mb-3">Headers sent with every request (e.g. Authorization, X-API-Key).</p>

        <div class="d-flex flex-wrap ga-2 mb-3">
          <v-chip
            v-for="key in headersStore.presetKeys"
            :key="key"
            :color="isPresetActive(key) ? 'primary' : undefined"
            :variant="isPresetActive(key) ? 'tonal' : 'outlined'"
            size="small"
            @click="headersStore.addPreset(key)"
          >
            {{ key }}
          </v-chip>
        </div>

        <div
          v-for="(entry, index) in headersStore.headers"
          :key="index"
          class="d-flex align-center ga-2 mb-2"
        >
          <v-text-field
            :model-value="entry.key"
            label="Key"
            density="compact"
            variant="outlined"
            hide-details
            class="flex-grow-1"
            @update:model-value="(v) => headersStore.updateHeader(index, { key: v })"
          />
          <v-text-field
            :model-value="entry.value"
            :type="visibleHeaderIndex === index ? 'text' : 'password'"
            label="Value"
            density="compact"
            variant="outlined"
            hide-details
            class="flex-grow-1"
            @update:model-value="(v) => headersStore.updateHeader(index, { value: v })"
          >
            <template #append-inner>
              <v-btn
                icon
                size="x-small"
                variant="text"
                @click="toggleHeaderVisible(index)"
              >
                <v-icon size="16">{{ visibleHeaderIndex === index ? 'visibility_off' : 'visibility' }}</v-icon>
              </v-btn>
            </template>
          </v-text-field>
          <v-btn
            icon
            variant="text"
            size="small"
            @click="headersStore.removeHeader(index)"
          >
            <v-icon size="18">close</v-icon>
          </v-btn>
        </div>

        <v-btn
          variant="tonal"
          size="small"
          prepend-icon="add"
          class="mt-2"
          @click="headersStore.addHeader()"
        >
          Add header
        </v-btn>
      </div>

      <v-divider class="mb-4" />

      <!-- OAuth 2.0 Flow -->
      <div class="mb-6">
        <div class="text-subtitle-2 mb-3">OAuth 2.0 Flow</div>

        <v-form
          ref="oauthFormRef"
          v-model="oauthFormValid"
          lazy-validation
          @submit.prevent="onAuthorizeConnection"
        >
          <v-text-field
            :model-value="oauthStore.config.clientId"
            label="Client ID"
            density="compact"
            variant="outlined"
            class="mb-3"
            :rules="[oauthRules.clientId]"
            @update:model-value="(v) => oauthStore.setConfig({ clientId: v ?? '' })"
          />

          <v-text-field
            :model-value="oauthStore.config.clientSecret"
            :type="showOAuthSecret ? 'text' : 'password'"
            label="Client Secret"
            density="compact"
            variant="outlined"
            class="mb-3"
            :rules="[oauthRules.clientSecret]"
            @update:model-value="(v) => oauthStore.setConfig({ clientSecret: v ?? '' })"
          >
            <template #append-inner>
              <v-btn
                icon
                size="x-small"
                variant="text"
                @click="showOAuthSecret = !showOAuthSecret"
              >
                <v-icon size="16">{{ showOAuthSecret ? 'visibility_off' : 'visibility' }}</v-icon>
              </v-btn>
            </template>
          </v-text-field>

          <v-text-field
            :model-value="oauthStore.config.authorizationUrl"
            label="Authorization URL"
            density="compact"
            variant="outlined"
            class="mb-3"
            :rules="[oauthRules.url]"
            @update:model-value="(v) => oauthStore.setConfig({ authorizationUrl: v ?? '' })"
          />

          <v-text-field
            :model-value="oauthStore.config.tokenUrl"
            label="Token URL"
            density="compact"
            variant="outlined"
            class="mb-3"
            :rules="[oauthRules.url]"
            @update:model-value="(v) => oauthStore.setConfig({ tokenUrl: v ?? '' })"
          />

          <v-text-field
            :model-value="redirectUrl"
            label="Redirect URL"
            density="compact"
            variant="outlined"
            hide-details
            readonly
            class="mb-3"
          />

          <v-text-field
            :model-value="oauthStore.config.scope"
            label="Scope"
            density="compact"
            variant="outlined"
            hide-details
            class="mb-3"
            placeholder="openid profile email"
            @update:model-value="(v) => oauthStore.setConfig({ scope: v ?? '' })"
          />

          <template v-if="!oauthStore.tokens">
            <v-btn
              block
              color="primary"
              type="submit"
              prepend-icon="verified_user"
              :disabled="!oauthFormValid"
            >
              Authorize Connection
            </v-btn>
          </template>
          <template v-else>
            <div class="d-flex flex-column ga-2">
              <span class="text-body-2 text-medium-emphasis mb-1">Signed in</span>
              <v-btn
                block
                variant="tonal"
                color="primary"
                prepend-icon="refresh"
                @click="onRefreshToken"
              >
                Refresh
              </v-btn>
              <v-btn
                block
                variant="tonal"
                color="error"
                prepend-icon="logout"
                @click="onSignOut"
              >
                Sign out
              </v-btn>
            </div>
          </template>
        </v-form>
      </div>
    </div>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
  import { useAgentPlaygroundStore } from '@/pages/playground/store/agentPlayground'
  import { useAgentConnectionStore } from '@/store/agentConnection'
  import { useAgentHeadersStore } from '@/store/agentHeaders'
  import { useAgentOAuthStore } from '@/store/agentOAuth'
  import { useSnackbarStore } from '@/store/snackbar'
  import { computed, ref } from 'vue'
  import { useDisplay } from 'vuetify'
  import type { VForm } from 'vuetify/components'

  const connectionStore = useAgentConnectionStore()
  const headersStore = useAgentHeadersStore()
  const oauthStore = useAgentOAuthStore()
  const snackbarStore = useSnackbarStore()
  const playgroundStore = useAgentPlaygroundStore()
  const { width } = useDisplay()

  const oauthFormRef = ref<VForm>()
  const oauthFormValid = ref(false)
  const visibleHeaderIndex = ref<number | null>(null)
  const showOAuthSecret = ref(false)

  const oauthRules = {
    clientId: (v: string) => {
      const value = (v ?? '').trim()
      if (!value) return 'Client ID is required'
      return true
    },
    clientSecret: (v: string) => {
      // Optional; no validation
      return true
    },
    url: (v: string) => {
      const value = (v ?? '').trim()
      if (!value) return 'This field is required'
      try {
        new URL(value)
        return true
      } catch {
        return 'Please enter a valid URL'
      }
    },
  }

  const drawerWidth = computed(() => {
    const w = width.value
    if (w >= 1920) return 500 // 2K and above
    if (w >= 1440) return 450 // Large desktop
    if (w >= 1024) return 400 // Desktop
    return 360 // Tablet and below
  })

  const toggleHeaderVisible = (index: number) => {
    visibleHeaderIndex.value = visibleHeaderIndex.value === index ? null : index
  }

  const isPresetActive = (key: string) => {
    return headersStore.headers.some((h) => (h.key || '').trim() === key)
  }

  const redirectUrl = computed(() => {
    if (typeof window === 'undefined') return ''
    return `${window.location.origin}/auth/callback`
  })

  const transportItems = [
    { title: 'gRPC', value: 'grpc' },
    { title: 'JSON-RPC', value: 'jsonrpc' },
  ] as const

  const onClose = () => {
    playgroundStore.setSidebarOpen(false)
  }

  const onRefreshToken = async () => {
    try {
      const ok = await oauthStore.refresh()
      if (ok) snackbarStore.success('Token refreshed')
      else snackbarStore.error('Refresh failed')
    } catch {
      snackbarStore.error('Refresh failed')
    }
  }

  const onSignOut = () => {
    oauthStore.signOut()
    snackbarStore.success('Signed out')
  }

  const onAuthorizeConnection = async () => {
    const validated = await oauthFormRef.value?.validate()
    if (!validated || !validated.valid) {
      const errorMessage = validated?.errors.map((err: { id: string | number; errorMessages: string[] }) => err.errorMessages.join(', ')).join(' ')
      snackbarStore.error(errorMessage ?? 'Please ensure all required fields are valid.')
      return
    }
    try {
      await oauthStore.authorize()
      snackbarStore.success('Authorization successful')
    } catch (e) {
      snackbarStore.error(e instanceof Error ? e.message : 'Authorization failed')
    }
  }
</script>
