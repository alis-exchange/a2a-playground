<template>
  <div
    class="text-part"
    v-html="renderedMarkdown"
  />
</template>

<script setup lang="ts">
  // Vue & Store Imports
  import type { Part } from '@local/a2a-js'
  import { marked } from 'marked'
  import type { Ref } from 'vue'
  import { computed } from 'vue'

  // Component Imports

  // Props
  interface Props {
    part: Part
  }

  const props = defineProps<Props>()

  // Define Stores

  // Define Emits

  // Refs

  // Computed
  const renderedMarkdown = computed(() => {
    const text = props.part.content?.case === 'text' ? (props.part.content.value ?? '') : ''
    if (!text || text.trim().length === 0) {
      return '&nbsp;'
    }

    try {
      // Render markdown to HTML
      const html = marked.parse(text) as string
      return html
    } catch {
      // Fallback to plain text if markdown parsing fails
      return text.replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/\n/g, '<br>')
    }
  })

  // Watchers

  // Functions
  // Configure marked for clean typography (matching Flutter app)
  marked.setOptions({
    breaks: true,
    gfm: true,
  })

  // Lifecycle Hooks
</script>

<style lang="scss" scoped>
  .text-part {
    font-size: 16px;
    line-height: 1.5;
    letter-spacing: 0;
    word-wrap: break-word;

    // Markdown styling to match Flutter app
    :deep(h1) {
      font-size: 32px;
      font-weight: bold;
      margin: 0.5em 0;
    }

    :deep(h2) {
      font-size: 28px;
      font-weight: bold;
      margin: 0.5em 0;
    }

    :deep(h3) {
      font-size: 24px;
      font-weight: bold;
      margin: 0.5em 0;
    }

    :deep(h4) {
      font-size: 20px;
      font-weight: 500;
      margin: 0.5em 0;
    }

    :deep(h5) {
      font-size: 18px;
      font-weight: 500;
      margin: 0.5em 0;
    }

    :deep(h6) {
      font-size: 16px;
      font-weight: 500;
      margin: 0.5em 0;
    }

    :deep(p) {
      margin: 0.5em 0;
    }

    :deep(ul),
    :deep(ol) {
      margin: 0.5em 0;
      padding-left: 1.5em;
    }

    :deep(li) {
      margin: 0.25em 0;
    }

    :deep(code) {
      font-family: monospace;
      font-size: 14px;
      background-color: rgba(0, 0, 0, 0.05);
      padding: 2px 4px;
      border-radius: 4px;
    }

    :deep(pre) {
      font-family: monospace;
      font-size: 14px;
      background-color: rgba(0, 0, 0, 0.05);
      padding: 8px;
      border-radius: 8px;
      overflow-x: auto;
    }

    :deep(blockquote) {
      border-left: 4px solid rgba(0, 0, 0, 0.2);
      padding-left: 1em;
      margin: 0.5em 0;
    }

    :deep(a) {
      color: inherit;
      text-decoration: underline;
    }
  }
</style>
