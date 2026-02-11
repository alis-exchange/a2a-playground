// Styles
import 'material-symbols'
import 'vuetify/styles'

// Material Design 3
//
// Since Vuetify 3 is lacking some of the md3 specifications,
// this section of imports aims to align it.
// Once the changes are implemented, these can be removed.
import LazyImage from '../../assets/lazy_img.png'
import '../../assets/md3-typography.css' // Override for md3 typography

// Vuetify
import type { ThemeDefinition } from 'vuetify'
import { createVuetify } from 'vuetify'
import { md3 } from 'vuetify/blueprints'
import { VBtn, VBtnGroup, VCard, VCardTitle, VChip, VFab, VIcon, VTab, VTabs, VTextField } from 'vuetify/components'
import { VList, VListItem } from 'vuetify/components/VList'
import * as directives from 'vuetify/directives'
import { aliases, md } from 'vuetify/iconsets/md'
import { VDateInput } from 'vuetify/labs/VDateInput'
import { VFileUpload, VFileUploadItem } from 'vuetify/labs/VFileUpload'
import { VStepperVertical, VStepperVerticalActions, VStepperVerticalItem } from 'vuetify/labs/VStepperVertical'
import { VVideo } from 'vuetify/labs/VVideo'

const lightTheme: ThemeDefinition = {
  dark: false,
  colors: {
    // Default colors
    anchor: '#000000',
    primary: '#006383',
    secondary: '#F02222',
    tertiary: '#C5E4CE',
    petrol: '#27323A',
    cyan: '#006383',
    twilight: '#3F2A6C',

    // Container colors
    background: '#F2F6F9',
    surface: '#ffffff',

    // State colors
    error: '#ba1a1a',
    running: '#199f19',
    failure: '#B00020',
    cancel: '#ba1a1a',
    run: '#199f19',
    unspecified: '#777777',
    pending: '#2196f3',
    new: '#2196f3',
    completed: '#199f19',

    // Surface Container
    surfaceContainerHighest: '#e6e8eb',
    surfaceContainerHigher: '#e9eef6',
    surfaceContainerHigh: '#eeeef1',
    surfaceContainer: '#ECF2F6',
    surfaceContainerLow: '#f7f9fc',
    surfaceContainerLowest: '#FFFFFF',

    secondarySurfaceContainer: '#EFEAF7',

    // On colors
    onPrimary: '#FFFFFF',
    'onPrimary--high': '#E0E0E0',
    'onPrimary--medium': '#A0A0A0',
    'onPrimary--disabled': '#6C6C6C',
    onBackground: '#777777',
    'onSurface--high': '#000000',
    'onSurface--medium': '#000000',
    'onSurface--disabled': '#000000',
    onError: '#FFFFFF',
  },
}

export default createVuetify({
  aliases: {
    // NAVIGATION RELATED ITEMS
    /** @description The primary navigation pane, used as a container for main navigation links. */
    XNavigationPane: VList,
    /** @description The primary action button, of which there should be maximum one, in the @see XNavigationPane. */
    XNavigationPaneFab: VBtn,
    /** @description A styled list item designed for use within @see XNavigationPane. */
    XListItem: VListItem,

    /** @description A container for tab-based navigation, often used for sub-sections. */
    XNavigationTabs: VTabs,
    /** @description An individually styled tab intended for use inside @see XNavigationTabs. */
    XNavigationTab: VTab,

    // MAIN CONTENT ON ANY PAGE
    /** @description The main content container for a page, styled as a primary card. */
    XMainCardContainer: VCard,
    /** @description A styled title for use within an @see XMainCardContainer. */
    XMainTitle: VCardTitle,

    // RESOURCE CARDS
    /** @description A card with consistent styling for displaying a single item in a resource list. */
    XResourceCard: VCard,
    /** @description The standard title component for an @see XResourceCard. */
    XResourceCardTitle: VCardTitle,

    // FILTERS THAT CAN BE APPLIED ON LIST VIEWS
    /** @description The standard title component for an @see XResourceCard. */
    XFilterContainer: VBtnGroup,
    /** @description Used as the unselected state within the @see XFilterContainer */
    XFilterItem: VBtn,
    /** @description Used as the selected state within the @see XFilterContainer */
    XFilterSelectedItem: VBtn,
    /** @description to remove the current filter state */
    XFilterRemoveBtn: VBtn,

    // MISC
    /** @description A pre-styled text field for global or top-level search functionality. */
    XSearchBar: VTextField,

    // Vuetify overrides
    VBtnIcon: VBtn,
    VCardOutlined: VCard,
    VChipIcon: VChip,
    VFabIcon: VFab,
    VPreIcon: VIcon,
    VPreFabIcon: VIcon,
    VTextFieldSearch: VTextField,
  },
  icons: {
    defaultSet: 'md',
    aliases,
    sets: {
      md,
    },
  },
  theme: {
    defaultTheme: 'lightTheme',
    themes: {
      lightTheme,
    },
  },
  defaults: {
    // We define a set of top level styles for alias components
    // starting with `x-` prefix to distinguish from the rest.

    // XNavigationPane is to be used for top level pages that have a navigation
    // pane on the left
    XNavigationPane: {
      class: ['x-navigation-pane', 'x-transition-width', 'bg-background', 'pl-3 pr-0', 'w-100', 'ml-n4'],
      nav: true,
      lines: 'one',
      density: 'default',

      // The primary action FAB that can be used in the navigation pane
      XNavigationPaneFab: {
        class: ['ml-2 mb-6 px-6 text-none font-weight-medium text-body-2'],
        height: '56px',
        color: 'primary',
        rounded: 'lg',
        elevation: 4,
        app: false,

        VIcon: {
          class: ['ml-n2', 'material-symbols-outlined'],
          size: 24,
        },
      },
      // The actual navigation elements
      XListItem: {
        class: ['x-list-item', 'text-start text-grey-darken-3 rounded-xl pl-4 mb-1'],
        color: 'petrol',
        activeClass: ['font-weight-bold'],
        rounded: 'xl',
        minHeight: '32px',
        density: 'default',
        style: ['font-size: 14px'],

        VAvatar: {
          size: 20,
          class: ['border-thin'],
        },
      },
    },
    XNavigationTabs: {
      backgroundColor: 'white',
      color: 'primary',
      grow: true,
      alignTabs: 'center',

      XNavigationTab: {
        activeColor: 'primary',

        VIcon: {
          class: ['mr-2', 'material-symbols-outlined'],
        },
      },
    },

    XMainCardContainer: {
      class: ['x-transition-width', 'justify-center ml-3 px-md-2 w-100'],
      rounded: 'xl',
      flat: true,

      XMainTitle: {
        class: ['text-md-h5 text-h6 mx-n2 px-6'],
      },
    },

    XFilterContainer: {
      density: 'compact',
      rounded: 'sm',
      variant: 'tonal',
    },
    XFilterItem: {
      class: ['pr-4 text-none'],
      VPreIcon: {
        class: ['mr-2', 'ml-n2', 'material-symbols-outlined'],
        size: 16,
      },
      variant: 'outlined',
      rounded: 'sm',
      color: 'grey-darken-3',
    },
    XFilterSelectedItem: {
      class: ['pr-4 text-none rounded-l-sm rounded-r-0'],
      VPreIcon: {
        class: ['mr-2', 'ml-n2', 'material-symbols-outlined'],
        size: 16,
      },
      variant: 'tonal',
      color: 'primary',
    },
    XFilterRemoveBtn: {
      class: ['text-none rounded-r-sm rounded-l-0'],
      VIcon: {
        class: ['material-symbols-outlined'],
        size: 16,
        style: ['margin-bottom: 2px'],
      },
      variant: 'tonal',
      color: 'primary',
      size: 'small',
      icon: 'close',
      style: ['margin-left: 1px'],
    },

    XSearchBar: {
      variant: 'solo',
      flat: true,
      bgColor: 'surfaceContainerHighest',
      density: 'compact',
      hideDetails: true,
      rounded: true,
      prependInnerIcon: 'search',
    },

    XResourceCard: {
      variant: 'flat',
      rounded: 'md',
      hover: true,
      color: 'background',
      class: ['pa-2 h-100'],

      XResourceCardTitle: {
        class: ['text-wrap font-weight-regular text-grey-darken-4'],
      },
    },

    VBreadcrumbs: {
      class: ['py-0 px-0'],
      density: 'compact',
    },
    VAppBar: {
      color: 'background',
    },
    VBottomNavigation: {
      VBtn: {
        class: ['text-label'],
        VIcon: {
          class: ['material-symbols-outlined'],
          color: 'petrol',
        },
      },
    },
    VNavigationDrawer: {
      VList: {
        class: ['justify-content-center', 'px-0'],
        nav: true,
      },
      VListItem: {
        class: ['text-center', 'text-petrol', 'text-label'],
      },
      VBtn: {
        class: ['px-2', 'my-1'],
        VIcon: {
          class: ['material-symbols-outlined'],
          color: 'petrol',
        },
        VPreIcon: {
          class: ['mr-2', 'ml-n2', 'material-symbols-outlined'],
          size: 16,
        },
      },
    },
    VBtn: {
      class: ['text-none px-6'],
      color: 'primary',
      rounded: 'xl',
      variant: 'tonal',
      VPreIcon: {
        class: ['mr-2', 'ml-n2', 'material-symbols-outlined'],
        size: 16,
      },
    },
    VBtnIcon: {
      color: 'primary',
      variant: 'tonal',
      icon: true,
    },
    VChip: {
      class: ['px-4'],
      rounded: 'xl',
    },
    VFab: {
      class: ['px-0 text-body-1 font-weight-medium'],
      height: '64px',
      color: 'primary',
      app: true,
      appear: true,
      offset: true,
      size: 'large',
      location: 'bottom end',
      VPreFabIcon: {
        class: ['mr-2', 'ml-n2', 'material-symbols-outlined'],
        size: '28px',
      },
      style: ['border-radius: 16px', 'margin-bottom: 32px', 'margin-left: 0px'],
    },
    VFabIcon: {
      class: ['px-0'],
      height: '64px',
      width: '8px',
      elevation: 3,
      color: 'primary',
      app: true,
      appear: true,
      offset: true,
      location: 'bottom end',
      VIcon: {
        class: ['material-symbols-outlined'],
        size: '28px',
      },
      style: ['border-radius: 16px', 'margin-bottom: 36px', 'margin-left: -4px'],
    },
    VChipIcon: {
      class: ['pl-3', 'pr-4'],
      rounded: 'xl',
      VIcon: {
        class: ['mr-2', 'material-symbols-outlined'],
        size: '16',
      },
    },
    VTextField: {
      variant: 'outlined',
      density: 'compact',
      color: 'primary',
      rounded: 'lg',
    },
    VTextFieldSearch: {
      variant: 'solo',
      flat: true,
      density: 'compact',
      hideDetails: true,
      rounded: true,
      prependInnerIcon: 'search',
    },
    VAlert: {
      class: ['w-100'],
      variant: 'tonal',
      density: 'compact',
    },
    VSelect: {
      variant: 'outlined',
      density: 'compact',
      color: 'primary',
      rounded: 'lg',
    },
    VProgressCircular: {
      indeterminate: true,
      color: 'primary',
    },
    VTextarea: {
      variant: 'outlined',
      density: 'compact',
      color: 'primary',
      rounded: 'lg',
    },
    VAutocomplete: {
      variant: 'outlined',
      density: 'compact',
      color: 'primary',
      rounded: 'lg',
    },
    VCard: {
      rounded: 'xl',
      flat: true,
    },
    VCardTitle: {
      class: ['text-wrap'],
    },
    VCardSubtitle: {
      class: ['text-wrap'],
    },
    VCardOutlined: {
      variant: 'outlined',
      style: ['border: 1px solid #e1e3e1;'],
      rounded: 'xl',
      flat: true,
    },
    VIcon: {
      class: ['material-symbols-outlined'],
      size: 20,
    },
    VImg: {
      lazySrc: LazyImage,
    },
    VSnackbar: {
      rounded: 'lg',
    },
    VTable: {
      density: 'comfortable',

      VChip: {
        size: 'small',

        VIcon: {
          class: ['material-symbols-outlined'],
          size: '16',
        },
      },
      VAvatar: {
        size: 'small',
      },
    },
    VDataTableVirtual: {
      density: 'comfortable',

      VChip: {
        size: 'small',

        VIcon: {
          class: ['material-symbols-outlined'],
          size: '16',
        },
      },
      VAvatar: {
        size: 'small',
      },
      VBtn: {
        size: `small`,
        VIcon: {
          class: ['material-symbols-outlined'],
          size: '16',
        },
      },
    },
    VTooltip: {
      location: 'bottom',
      openDelay: '200',
      maxWidth: '240',
    },
  },
  blueprint: md3,
  components: { VDateInput, VStepperVertical, VStepperVerticalItem, VStepperVerticalActions, VFileUpload, VFileUploadItem, VVideo },
  directives,
})
