import { registerPlugins } from '@/plugins'
import { createApp } from 'vue'
import App from './App.vue'

// Styles
import 'unfonts.css'
import './assets/tiptap-styles.css'

const app = createApp(App)

registerPlugins(app)

app.mount('#app')
