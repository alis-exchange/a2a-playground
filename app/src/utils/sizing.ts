import { computed, onMounted, onUnmounted, type Ref, ref } from 'vue'
import { useDisplay } from 'vuetify'
import { VCard } from 'vuetify/lib/components/index.mjs'

/**
 * The desired margin to maintain at the bottom of the card, in pixels.
 */
const CARD_BOTTOM_MARGIN = 16

/**
 * The height of the bottom navigation bar on mobile devices, in pixels.
 * This needs to be subtracted from the available height on small screens.
 */
const NAVIGATION_BOTTOM_HEIGHT = 56

/**
 * A Vue composable that calculates the dynamic height for a `VCard` component
 * to make it fill the remaining vertical space on the screen.
 *
 * @returns An object containing:
 *  - `appPanelCardRef`: A ref to be attached to the `V-Card` component instance.
 *  - `appPanelCardHeight`: A computed property with the calculated height in pixels.
 *
 * @example
 * <template>
 *   <v-card :ref="appPanelCardRef" :style="{ height: appPanelCardHeight + 'px' }">
 *     ...
 *   </v-card>
 * </template>
 *
 * <script setup>
 *   import { useAppPanelCardHeight } from './utils/sizing';
 *   const { appPanelCardRef, appPanelCardHeight } = useAppPanelCardHeight();
 * <\/script>
 */
export function useAppPanelCardHeight() {
  const { height, width, smAndDown } = useDisplay()
  const appPanelCardRef: Ref<VCard | null> = ref(null)

  /**
   * A trigger to manually force re-computation of the card height.
   * Vue's computed properties only re-evaluate when their reactive dependencies change.
   * However, the card's top position can be affected by DOM changes outside of Vue's
   * reactivity system (e.g., a banner appearing). We use a MutationObserver to detect
   * such changes and increment this trigger, forcing `appPanelCardHeight` to recalculate.
   */
  const layoutTrigger = ref(0)

  const appPanelCardHeight = computed(() => {
    // By accessing these reactive properties, we ensure the computed property
    // re-evaluates whenever the window width, height, or our manual trigger changes.
    const _w = width.value
    const _h = height.value
    const _trigger = layoutTrigger.value

    if (appPanelCardRef.value?.$el) {
      // The core calculation:
      // 1. Get the card's distance from the top of the viewport.
      const cardTopPosition = appPanelCardRef.value.$el.getBoundingClientRect().top

      // 2. Calculate the available height from the card's top to the bottom of the viewport.
      let availableHeight = _h - cardTopPosition - CARD_BOTTOM_MARGIN

      // 3. On mobile, subtract the height of the bottom navigation bar.
      if (smAndDown.value) {
        availableHeight -= NAVIGATION_BOTTOM_HEIGHT
      }

      // 4. Ensure the calculated height is never negative.
      return Math.max(0, availableHeight)
    }

    // Return 0 if the card reference is not yet available.
    return 0
  })

  // A MutationObserver is used to watch for changes in the DOM that might affect
  // the card's position. For example, if a notification banner is dynamically added
  // at the top of the page, it will push our card down, and we need to recalculate its height.
  let observer: MutationObserver | null = null

  onMounted(() => {
    // We use requestAnimationFrame to ensure the DOM is fully painted and settled
    // before we attach the observer.
    requestAnimationFrame(() => {
      // We observe the card's direct parent element. Any change to the children
      // of the parent (e.g., adding a sibling element above the card) will be detected.
      const parentElement = appPanelCardRef.value?.$el?.parentElement
      if (parentElement) {
        observer = new MutationObserver(() => {
          // When a mutation is detected, we increment the layoutTrigger.
          // This makes the `appPanelCardHeight` computed property re-evaluate
          // its value, adjusting the card's height to the new layout.
          layoutTrigger.value++
        })

        // Configure the observer to watch for additions or removals of child elements
        // within the parent, including those in the entire subtree.
        observer.observe(parentElement, {
          childList: true, // Watch for direct children being added or removed.
          subtree: true, // Watch for changes in all descendants.
        })
      }
    })
  })

  // It's crucial to disconnect the observer when the component is unmounted
  // to prevent memory leaks.
  onUnmounted(() => {
    if (observer) {
      observer.disconnect()
    }
  })

  return {
    appPanelCardRef,
    appPanelCardHeight,
  }
}
