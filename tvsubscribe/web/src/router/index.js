import Config from '../views/Config.vue'
import Subscribe from '../views/Subscribe.vue'

const routes = [
  {
    path: '/',
    redirect: '/config'
  },
  {
    path: '/config',
    name: 'Config',
    component: Config
  },
  {
    path: '/subscribe',
    name: 'Subscribe',
    component: Subscribe
  }
]

export default routes