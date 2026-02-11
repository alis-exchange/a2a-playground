<script setup lang="ts">
  import DetailsDrawer from '../DetailsDrawer.vue'
</script>

<template>
  <v-app>
    <v-main>
      <slot />
    </v-main>
    <DetailsDrawer />
  </v-app>
</template>
