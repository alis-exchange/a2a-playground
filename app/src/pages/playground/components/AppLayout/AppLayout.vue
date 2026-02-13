<script setup lang="ts">
  import DetailsDrawer from '../DetailsDrawer.vue'
  import HeadersDrawer from '../HeadersDrawer.vue'
</script>

<template>
  <v-app>
    <v-main>
      <slot />
    </v-main>
    <DetailsDrawer />
    <HeadersDrawer />
  </v-app>
</template>
