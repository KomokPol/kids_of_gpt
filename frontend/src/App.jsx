import { useEffect } from 'react'
import Header from './ui/Header/Header.jsx'
import AppRouter from './router/index.jsx'
import useUserStore from './store/useUserStore.js'
import styles from './App.module.css'

export default function App() {
  const fetchUser = useUserStore(s => s.fetchUser)

  // Загружаем данные пользователя при старте приложения
  useEffect(() => {
    fetchUser()
  }, [fetchUser])

  return (
    <div className={styles.app}>
      <Header />
      <main className={styles.main}>
        <AppRouter />
      </main>
    </div>
  )
}
