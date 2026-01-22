import { createApp, watch } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import i18n from './i18n'

const app = createApp(App)

app.use(router)
app.use(i18n)

const { locale, t } = i18n.global

const updateTitle = () => {
  document.title = t('appTitle')
}

watch(locale, () => {
  updateTitle()
}, { immediate: true })

app.mount('#app')
